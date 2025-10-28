package component

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

var _ resource.Resource = &componentResource{}
var _ resource.ResourceWithImportState = &componentResource{}

type componentResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &componentResource{}
}

func (r *componentResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (r *componentResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst Component resource",
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Component name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Component type",
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Component version",
				Required:            true,
			},
			"spec": schema.StringAttribute{
				MarkdownDescription: "Dapr Component Metadata in YAML format",
				Required:            true,
			},
			"secret_store": schema.StringAttribute{
				MarkdownDescription: "Secret store for secret reference resolution",
				Optional:            true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "App IDs that can access this component",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the component",
				Computed:            true,
			},
		},
	}
}

func (r *componentResource) Configure(ctx context.Context,
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

func (r *componentResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "creating component")

	component := &client.DaprComponent{
		ApiVersion: lo.ToPtr("dapr.io/v1alpha1"),
		Kind:       lo.ToPtr("Component"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.DaprComponentSpec{
			Type:    lo.ToPtr(model.GetType()),
			Version: lo.ToPtr(model.GetVersion()),
		},
	}

	// Convert YAML spec to API metadata array
	if !model.Spec.IsNull() && !model.Spec.IsUnknown() {
		metadata, err := toAPISpec(ctx, model.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Spec",
				fmt.Sprintf("error parsing spec YAML: %s", err))
			return
		}
		component.Spec.Metadata = metadata
	}

	// Set auth (secret_store)
	if !model.SecretStore.IsNull() && !model.SecretStore.IsUnknown() {
		component.Auth = &client.DaprComponentAuth{
			SecretStore: lo.ToPtr(model.GetSecretStore()),
		}
	}

	// Set scopes
	scopes := toAPIScopes(ctx, model.Scopes)
	if len(scopes) > 0 {
		component.Scopes = &scopes
	}

	if err := r.client.CreateComponent(ctx, model.GetProjectName(), component); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating component: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading component after create: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *componentResource) Read(ctx context.Context,
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
			fmt.Sprintf("error reading component: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *componentResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "updating component")

	component := &client.DaprComponent{
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.DaprComponentSpec{
			Type:    lo.ToPtr(model.GetType()),
			Version: lo.ToPtr(model.GetVersion()),
		},
	}

	// Convert YAML spec to API metadata array
	if !model.Spec.IsNull() && !model.Spec.IsUnknown() {
		metadata, err := toAPISpec(ctx, model.Spec)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Spec",
				fmt.Sprintf("error parsing spec YAML: %s", err))
			return
		}
		component.Spec.Metadata = metadata
	}

	// Set auth (secret_store)
	if !model.SecretStore.IsNull() && !model.SecretStore.IsUnknown() {
		component.Auth = &client.DaprComponentAuth{
			SecretStore: lo.ToPtr(model.GetSecretStore()),
		}
	}

	// Set scopes
	scopes := toAPIScopes(ctx, model.Scopes)
	if len(scopes) > 0 {
		scopesTyped := scopes
		component.Scopes = &scopesTyped
	}

	if err := r.client.UpdateComponent(ctx, model.GetProjectName(), model.GetName(), component); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating component: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading component after update: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *componentResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "deleting component")

	if err := r.client.DeleteComponent(ctx, model.GetProjectName(), model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Info(ctx, "component not found, considering it deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting component: %s", err))
		return
	}
}

func (r *componentResource) ImportState(ctx context.Context,
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
			tflog.Debug(ctx, "component not found", map[string]interface{}{
				"project_name": model.GetProjectName(),
				"name":         model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported component: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
