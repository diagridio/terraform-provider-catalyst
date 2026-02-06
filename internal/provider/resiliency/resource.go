package resiliency

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

	"github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	diagrid_errors "github.com/diagridio/cloudgrid/sdk/go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ resource.Resource = &resiliencyResource{}
var _ resource.ResourceWithImportState = &resiliencyResource{}

type resiliencyResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &resiliencyResource{}
}

func (r *resiliencyResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_resiliency"
}

func (r *resiliencyResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst Resiliency resource",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resiliency name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"spec": resiliencySpecAttribute(true, false),
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

func (r *resiliencyResource) Configure(ctx context.Context,
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

func (r *resiliencyResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "creating resiliency")

	resiliency := &client.DaprResiliency{
		ApiVersion: lo.ToPtr("dapr.io/v1alpha1"),
		Kind:       lo.ToPtr("Resiliency"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
	}

	// Set scopes
	resiliency.Scopes = toAPIScopes(ctx, model.Scopes)

	spec, err := expandResiliencySpec(ctx, model.ensureSpec())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Spec",
			fmt.Sprintf("error constructing spec: %s", err))
		return
	}
	resiliency.Spec = spec

	if err := r.client.CreateResiliency(ctx, model.GetProjectID(), resiliency); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating resiliency: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading resiliency after create: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *resiliencyResource) Read(ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading resiliency: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *resiliencyResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "updating resiliency")

	resiliency := &client.DaprResiliency{
		ApiVersion: lo.ToPtr("dapr.io/v1alpha1"),
		Kind:       lo.ToPtr("Resiliency"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
	}

	// Set scopes
	resiliency.Scopes = toAPIScopes(ctx, model.Scopes)

	spec, err := expandResiliencySpec(ctx, model.ensureSpec())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Spec",
			fmt.Sprintf("error constructing spec: %s", err))
		return
	}
	resiliency.Spec = spec

	if err := r.client.UpdateResiliency(ctx, model.GetProjectID(), model.GetName(), resiliency); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating resiliency: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading resiliency after update: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *resiliencyResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "deleting resiliency")

	if err := r.client.DeleteResiliency(ctx, model.GetProjectID(), model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Info(ctx, "resiliency not found, considering it deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting resiliency: %s", err))
		return
	}
}

func (r *resiliencyResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Parse the import ID (format: "project_id/name")
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'project_id/name', got: %s", req.ID))
		return
	}

	model := NewModel()
	model.SetProjectID(parts[0])
	model.SetName(parts[1])

	if err := read(ctx, r.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "resiliency not found", map[string]interface{}{
				"project_id": model.GetProjectID(),
				"name":       model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported resiliency: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
