package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ resource.Resource = &configurationResource{}
var _ resource.ResourceWithImportState = &configurationResource{}

type configurationResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &configurationResource{}
}

func (r *configurationResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (r *configurationResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst Configuration resource",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Configuration name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"spec": configurationSpecAttribute(true, false),
			"status": schema.StringAttribute{
				MarkdownDescription: "Status",
				Computed:            true,
			},
		},
	}
}

func (r *configurationResource) Configure(ctx context.Context,
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

func (r *configurationResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "creating configuration")

	// Save the planned spec to preserve user-specified values
	plannedSpec := model.Spec

	configuration := &client.DaprConfiguration{
		ApiVersion: lo.ToPtr("dapr.io/v1alpha1"),
		Kind:       lo.ToPtr("Configuration"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
	}

	spec, err := expandConfigurationSpec(ctx, model.ensureSpec())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Spec",
			fmt.Sprintf("error constructing spec: %s", err))
		return
	}
	configuration.Spec = spec

	if err := r.client.CreateConfiguration(ctx, model.GetProjectID(), configuration); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating configuration: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading configuration after create: %s", err))
		return
	}

	// Merge planned spec values with read values to preserve user configuration
	mergeSpecWithPlanned(model.Spec, plannedSpec)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *configurationResource) Read(ctx context.Context,
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
			fmt.Sprintf("error reading configuration: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *configurationResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "updating configuration")

	// Save the planned spec to preserve user-specified values
	plannedSpec := model.Spec

	configuration := &client.DaprConfiguration{
		ApiVersion: lo.ToPtr("dapr.io/v1alpha1"),
		Kind:       lo.ToPtr("Configuration"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
	}

	spec, err := expandConfigurationSpec(ctx, model.ensureSpec())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Spec",
			fmt.Sprintf("error constructing spec: %s", err))
		return
	}
	configuration.Spec = spec

	if err := r.client.UpdateConfiguration(ctx, model.GetProjectID(), model.GetName(), configuration); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating configuration: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading configuration after update: %s", err))
		return
	}

	// Merge planned spec values with read values to preserve user configuration
	mergeSpecWithPlanned(model.Spec, plannedSpec)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *configurationResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "deleting configuration")

	if err := r.client.DeleteConfiguration(ctx, model.GetProjectID(), model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Info(ctx, "configuration not found, considering it deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting configuration: %s", err))
		return
	}
}

func (r *configurationResource) ImportState(ctx context.Context,
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
			tflog.Debug(ctx, "configuration not found", map[string]interface{}{
				"project_id": model.GetProjectID(),
				"name":       model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported configuration: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
