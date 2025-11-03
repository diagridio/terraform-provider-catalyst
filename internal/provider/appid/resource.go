package appid

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

var _ resource.Resource = &appidResource{}
var _ resource.ResourceWithImportState = &appidResource{}

type appidResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &appidResource{}
}

func (r *appidResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_appid"
}

func (r *appidResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst AppID resource",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "AppID name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_config": schema.StringAttribute{
				MarkdownDescription: "App configuration name",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol (e.g., 'http', 'grpc')",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_token_revision": schema.Int64Attribute{
				MarkdownDescription: "API token revision",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the AppID",
				Computed:            true,
			},
			"app_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Application endpoint configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Application endpoint URL",
						Optional:            true,
					},
					"token": schema.StringAttribute{
						MarkdownDescription: "Authentication token",
						Optional:            true,
						Sensitive:           true,
					},
					"token_header": schema.StringAttribute{
						MarkdownDescription: "Header name for the authentication token",
						Optional:            true,
					},
					"client_timeout_seconds": schema.Int64Attribute{
						MarkdownDescription: "Client timeout in seconds",
						Optional:            true,
					},
				},
			},
			"health_check": schema.SingleNestedAttribute{
				MarkdownDescription: "Health check configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						MarkdownDescription: "Health check path",
						Optional:            true,
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether the probe is enabled",
						Optional:            true,
					},
					"failure_threshold": schema.Int64Attribute{
						MarkdownDescription: "Number of failures before marking as unhealthy",
						Optional:            true,
					},
					"interval_seconds": schema.Int64Attribute{
						MarkdownDescription: "Interval between probe checks in seconds",
						Optional:            true,
					},
					"timeout_ms": schema.Int64Attribute{
						MarkdownDescription: "Timeout for each probe check in milliseconds",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *appidResource) Configure(ctx context.Context,
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

func (r *appidResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "creating appid")

	// Store original plan values for optional computed fields
	originalProtocol := model.Protocol
	originalAppConfig := model.AppConfig

	appid := &client.AppIdentity{
		ApiVersion: lo.ToPtr("cra.diagrid.io/v1beta1"),
		Kind:       lo.ToPtr("AppIdentity"),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.AppIdentitySpec{},
	}

	// Set optional fields
	if !model.AppConfig.IsNull() && !model.AppConfig.IsUnknown() {
		appid.Spec.AppConfig = lo.ToPtr(model.GetAppConfig())
	}

	if !model.Protocol.IsNull() && !model.Protocol.IsUnknown() {
		appid.Spec.Protocol = lo.ToPtr(model.GetProtocol())
	}

	if !model.ApiTokenRevision.IsNull() && !model.ApiTokenRevision.IsUnknown() {
		revision := int(model.GetApiTokenRevision())
		appid.Spec.ApiTokenRevision = &revision
	}

	// Set app endpoint
	appid.Spec.AppEndpoint = toAPIAppEndpoint(ctx, model.AppEndpoint)

	// Set health check
	appid.Spec.HealthCheck = toAPIHealthCheck(ctx, model.HealthCheck)

	if err := r.client.CreateAppId(ctx, model.GetProjectID(), appid); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error creating appid: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading appid after create: %s", err))
		return
	}

	// Preserve null state for optional computed fields if they weren't explicitly set
	if originalProtocol.IsNull() {
		model.Protocol = originalProtocol
	}
	if originalAppConfig.IsNull() {
		model.AppConfig = originalAppConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *appidResource) Read(ctx context.Context,
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
			fmt.Sprintf("error reading appid: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *appidResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "updating appid")

	// Store original plan values for optional computed fields
	originalProtocol := model.Protocol
	originalAppConfig := model.AppConfig

	appid := &client.AppIdentity{
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.AppIdentitySpec{},
	}

	// Set optional fields
	if !model.AppConfig.IsNull() && !model.AppConfig.IsUnknown() {
		appid.Spec.AppConfig = lo.ToPtr(model.GetAppConfig())
	}

	if !model.Protocol.IsNull() && !model.Protocol.IsUnknown() {
		appid.Spec.Protocol = lo.ToPtr(model.GetProtocol())
	}

	if !model.ApiTokenRevision.IsNull() && !model.ApiTokenRevision.IsUnknown() {
		revision := int(model.GetApiTokenRevision())
		appid.Spec.ApiTokenRevision = &revision
	}

	// Set app endpoint
	appid.Spec.AppEndpoint = toAPIAppEndpoint(ctx, model.AppEndpoint)

	// Set health check
	appid.Spec.HealthCheck = toAPIHealthCheck(ctx, model.HealthCheck)

	if err := r.client.UpdateAppId(ctx, model.GetProjectID(), model.GetName(), appid); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error updating appid: %s", err))
		return
	}

	if err := read(ctx, r.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading appid after update: %s", err))
		return
	}

	// Preserve null state for optional computed fields if they weren't explicitly set
	if originalProtocol.IsNull() {
		model.Protocol = originalProtocol
	}
	if originalAppConfig.IsNull() {
		model.AppConfig = originalAppConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *appidResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Log(ctx, "deleting appid")

	if err := r.client.DeleteAppId(ctx, model.GetProjectID(), model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Info(ctx, "appid not found, considering it deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error deleting appid: %s", err))
		return
	}
}

func (r *appidResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import ID format: "project_id/name"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: project_id/name, got: %s", req.ID),
		)
		return
	}

	model := NewModel()
	model.SetProjectID(parts[0])
	model.SetName(parts[1])

	if err := read(ctx, r.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "appid not found", map[string]interface{}{
				"project_id": model.GetProjectID(),
				"name":       model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported appid: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
