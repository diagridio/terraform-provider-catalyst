package serviceaccount

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"

	"github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	diagrid_errors "github.com/diagridio/cloudgrid/sdk/go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
)

var _ resource.Resource = &serviceAccountResource{}
var _ resource.ResourceWithImportState = &serviceAccountResource{}

type serviceAccountResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &serviceAccountResource{}
}

func (s *serviceAccountResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (s *serviceAccountResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalyst service account resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Service account name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Service account description",
				Required:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "Service account owner",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Service account role",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Service account email",
				Computed:            true,
			},
		},
	}
}

func (s *serviceAccountResource) Configure(ctx context.Context,
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
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	s.client = providerData.Client
}

func (s *serviceAccountResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating service account",
		map[string]interface{}{
			"name":        model.GetName(),
			"description": model.GetDescription(),
			"owner":       model.GetOwner(),
			"role":        model.GetRole(),
		})

	serviceAccount := &client.ServiceAccount{
		ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
		Kind:       lo.ToPtr(catalyst.KindServiceAccount),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: client.ServiceAccountSpec{
			Description: model.GetDescription(),
			Owner:       model.GetOwner(),
			Role:        model.GetRole(),
		},
		Status: &client.ServiceAccountStatus{},
	}

	if err := s.client.CreateServiceAccount(ctx, serviceAccount); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error creating service account: %s", err))
		return
	}

	if err := helpers.WaitUntil(ctx,
		func(ctx context.Context) (bool, error) {
			serviceAccount, err := s.client.GetServiceAccount(ctx, model.GetName())
			if err != nil {
				return false, fmt.Errorf("error getting service account: %w", err)
			}

			if serviceAccount.Status != nil &&
				serviceAccount.Status.Email != nil {
				return true, nil
			}

			return false, nil
		}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting created service account: %s", err))
		return
	}

	if err := read(ctx, s.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading created service account: %s", err))
		return
	}

	tflog.Debug(ctx, "created service account", map[string]interface{}{
		"name": model.GetName(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (s *serviceAccountResource) Read(ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := read(ctx, s.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "service account not found", map[string]interface{}{
				"name": model.GetName(),
			})

			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading service account: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (s *serviceAccountResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccount, err := s.client.GetServiceAccount(ctx, model.GetName())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting service account: %s", err))
		return
	}

	serviceAccount.Spec.Description = model.GetDescription()
	serviceAccount.Spec.Owner = model.GetOwner()
	serviceAccount.Spec.Role = model.GetRole()

	tflog.Debug(ctx, "updating service account", map[string]interface{}{
		"serviceAccount": serviceAccount,
	})

	if err := s.client.UpdateServiceAccount(ctx, model.GetName(), serviceAccount); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error updating service account: %s", err))
		return
	}

	if err := read(ctx, s.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading updated service account: %s", err))
		return
	}

	tflog.Debug(ctx, "updated service account", map[string]interface{}{
		"model": model.String(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (s *serviceAccountResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting service account",
		map[string]interface{}{
			"name": model.GetName(),
		})

	if err := s.client.DeleteServiceAccount(ctx, model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "service account to delete not found", map[string]interface{}{
				"name": model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error deleting service account: %s", err))
		return
	}

	tflog.Debug(ctx, "deleted service account",
		map[string]interface{}{
			"name": model.GetName(),
		})
}

func (s *serviceAccountResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	model := NewModel()
	model.SetName(req.ID)

	if err := read(ctx, s.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			resp.Diagnostics.AddError("Resource Not Found",
				fmt.Sprintf("service account %s not found during import", model.GetName()))
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported service account: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
