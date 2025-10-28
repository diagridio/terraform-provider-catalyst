package subscription

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ resource.Resource = &subscriptionResource{}
var _ resource.ResourceWithImportState = &subscriptionResource{}

type subscriptionResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &subscriptionResource{}
}

func (r *subscriptionResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *subscriptionResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst Subscription resource",
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subscription name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pubsub_name": schema.StringAttribute{
				MarkdownDescription: "PubSub name",
				Required:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Topic name",
				Required:            true,
			},
			"spec": schema.StringAttribute{
				MarkdownDescription: "Dapr Subscription spec in YAML format",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					YAMLEquivalence(),
				},
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "Scopes",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status",
				Computed:            true,
			},
		},
	}
}

func (r *subscriptionResource) Configure(ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerData.Client
}

func (r *subscriptionResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "creating subscription")

	subscription := &client.DaprSubscription{
		ApiVersion: lo.ToPtr("dapr.io/v2alpha1"),
		Kind:       lo.ToPtr("Subscription"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.DaprSubscriptionSpec{
			Pubsubname: lo.ToPtr(model.GetPubsubName()),
			Topic:      lo.ToPtr(model.GetTopic()),
		},
	}

	// Convert YAML spec to API spec (merges with existing spec)
	if !model.Spec.IsNull() && !model.Spec.IsUnknown() {
		specFromYAML, err := toAPISpec(ctx, model.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Spec",
				fmt.Sprintf("error parsing spec YAML: %s", err))
			return
		}
		// Merge YAML spec fields with base spec (pubsubname and topic already set)
		if specFromYAML.Routes != nil {
			subscription.Spec.Routes = specFromYAML.Routes
		}
		if specFromYAML.DeadLetterTopic != nil {
			subscription.Spec.DeadLetterTopic = specFromYAML.DeadLetterTopic
		}
		if specFromYAML.BulkSubscribe != nil {
			subscription.Spec.BulkSubscribe = specFromYAML.BulkSubscribe
		}
		if specFromYAML.Metadata != nil {
			subscription.Spec.Metadata = specFromYAML.Metadata
		}
		if specFromYAML.Dynamic != nil {
			subscription.Spec.Dynamic = specFromYAML.Dynamic
		}
	}

	// Set scopes
	subscription.Scopes = toAPIScopes(ctx, model.Scopes)

	if err := r.client.CreateSubscription(ctx, model.GetProjectName(), subscription); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating subscription: %s", err))
		return
	}

	// Save the original spec YAML from the plan before reading
	originalSpec := model.Spec

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading subscription after create: %s", err))
		return
	}

	// Restore the original spec YAML to avoid formatting differences
	if !originalSpec.IsNull() && !originalSpec.IsUnknown() {
		model.Spec = originalSpec
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *subscriptionResource) Read(ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the original spec YAML from state before reading
	originalSpec := model.Spec

	if err := read(ctx, r.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading subscription: %s", err))
		return
	}

	// Restore the original spec YAML to avoid formatting differences
	if !originalSpec.IsNull() && !originalSpec.IsUnknown() {
		model.Spec = originalSpec
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *subscriptionResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "updating subscription")

	subscription := &client.DaprSubscription{
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.DaprSubscriptionSpec{
			Pubsubname: lo.ToPtr(model.GetPubsubName()),
			Topic:      lo.ToPtr(model.GetTopic()),
		},
	}

	// Convert YAML spec to API spec (merges with existing spec)
	if !model.Spec.IsNull() && !model.Spec.IsUnknown() {
		specFromYAML, err := toAPISpec(ctx, model.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Spec",
				fmt.Sprintf("error parsing spec YAML: %s", err))
			return
		}
		// Merge YAML spec fields with base spec
		if specFromYAML.Routes != nil {
			subscription.Spec.Routes = specFromYAML.Routes
		}
		if specFromYAML.DeadLetterTopic != nil {
			subscription.Spec.DeadLetterTopic = specFromYAML.DeadLetterTopic
		}
		if specFromYAML.BulkSubscribe != nil {
			subscription.Spec.BulkSubscribe = specFromYAML.BulkSubscribe
		}
		if specFromYAML.Metadata != nil {
			subscription.Spec.Metadata = specFromYAML.Metadata
		}
		if specFromYAML.Dynamic != nil {
			subscription.Spec.Dynamic = specFromYAML.Dynamic
		}
	}

	// Set scopes
	subscription.Scopes = toAPIScopes(ctx, model.Scopes)

	if err := r.client.UpdateSubscription(ctx, model.GetProjectName(), model.GetName(), subscription); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating subscription: %s", err))
		return
	}

	// Save the original spec YAML from the plan before reading
	originalSpec := model.Spec

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading subscription after update: %s", err))
		return
	}

	// Restore the original spec YAML to avoid formatting differences
	if !originalSpec.IsNull() && !originalSpec.IsUnknown() {
		model.Spec = originalSpec
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *subscriptionResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "deleting subscription")

	if err := r.client.DeleteSubscription(ctx, model.GetProjectName(), model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Info(ctx, "subscription not found, considering it deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting subscription: %s", err))
		return
	}
}

func (r *subscriptionResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import ID format: "project_name/name"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: project_name/name, got: %s", req.ID),
		)
		return
	}

	model := NewModel()
	model.SetProjectName(parts[0])
	model.SetName(parts[1])

	if err := read(ctx, r.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "subscription not found", map[string]interface{}{
				"project_name": model.GetProjectName(),
				"name":         model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported subscription: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
