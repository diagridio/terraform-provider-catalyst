package region

import (
	"context"
	"fmt"
	"regexp"

	"github.com/samber/lo"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &regionResource{}
var _ resource.ResourceWithImportState = &regionResource{}

var ingressRegex = regexp.MustCompile(`^https?://\*\.[^:]+:\d+$`)

// regionResource defines the resource implementation.
type regionResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &regionResource{}
}

func (p *regionResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

func (p *regionResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Catalyst region resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Region name",
				Required:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Region host",
				Optional:            true,
			},
			"ingress": schema.StringAttribute{
				MarkdownDescription: "Region Ingress provided by user; canonicalized by API",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						ingressRegex,
						"must start with http://*. or https://*., contain a hostname, and include a port",
					),
				},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region location",
				Optional:            true,
			},
			"join_token": schema.StringAttribute{
				MarkdownDescription: "Join token for the region",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Region type",
				Computed:            true,
			},
			"connected": schema.BoolAttribute{
				MarkdownDescription: "Whether the region is connected",
				Computed:            true,
			},
		},
	}
}

func (p *regionResource) Configure(ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
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

	p.client = providerData.Client
}

func (p *regionResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating private region",
		map[string]interface{}{
			"name":     model.GetName(),
			"host":     model.GetHost(),
			"ingress":  model.GetIngress(),
			"location": model.GetLocation(),
		})

	region := &client.Region{
		ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
		Kind:       lo.ToPtr(catalyst.KindRegion),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.RegionSpec{
			Host:     lo.ToPtr(model.GetHost()),
			Ingress:  lo.ToPtr(model.GetIngress()),
			Location: lo.ToPtr(model.GetLocation()),
		},
		Status: &client.RegionStatus{},
	}
	joinToken, err := p.client.CreateRegion(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error creating region: %s", err))
		return
	}

	// wait until region is created and in ready status
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		region, err = p.client.GetRegion(ctx, model.GetName())
		if err != nil {
			return false, fmt.Errorf("Error getting region: %w", err)
		}

		if region.Status != nil &&
			region.Status.Status != nil &&
			*region.Status.Status == "ready" {
			return true, nil
		}

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting region: %s", err))
		return
	}

	tflog.Debug(ctx, "created region",
		map[string]interface{}{
			"id":   region.Metadata.Uid,
			"name": *region.Metadata.Name,
		})

	// read back into the model
	if err := read(ctx, p.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading created region: %s", err))
		return
	}

	// Set the join token in the model, this is the only place we set it
	model.SetJoinToken(joinToken)

	tflog.Debug(ctx, "storing region model after creation",
		map[string]interface{}{
			"model": model.String(),
		})

	// Save model into Terraform state
	diagnostics := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diagnostics...)
}

func (p *regionResource) Read(ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	model := NewModel()

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := read(ctx, p.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "region not found",
				map[string]interface{}{
					"name": model.GetName(),
				})

			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading region: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (p *regionResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := p.client.GetRegion(ctx, model.GetName())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting region: %s", err))
		return
	}

	// scrub clusters from region, we're not allowed to update them
	region.Spec.Clusters = nil
	// same for region type
	region.Spec.Type = nil
	region.Spec.Host = lo.ToPtr(model.GetHost())
	region.Spec.Ingress = lo.ToPtr(model.GetIngress())
	region.Spec.Location = lo.ToPtr(model.GetLocation())

	tflog.Debug(ctx, "updating region",
		map[string]interface{}{
			"model": model.String(),
		})

	if err := p.client.UpdateRegion(ctx, region); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error updating region: %s", err))
		return
	}

	// wait until region is updated and in ready status
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		region, err := p.client.GetRegion(ctx, model.GetName())
		if err != nil {
			return false, fmt.Errorf("Error getting region: %w", err)
		}

		if region.Status != nil &&
			region.Status.Status != nil &&
			*region.Status.Status == "ready" {
			return true, nil
		}

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting region: %s", err))
		return
	}

	if err := read(ctx, p.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading updated region: %s", err))
		return
	}

	tflog.Debug(ctx, "updated region",
		map[string]interface{}{
			"model": model.String(),
		})

	// Save updated data into Terraform state
	diagnostics := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diagnostics...)
}

func (p *regionResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting region",
		map[string]interface{}{
			"name": model.GetName(),
		})

	if err := p.client.DeleteRegion(ctx, model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "region to delete not found",
				map[string]interface{}{
					"name": model.GetName(),
				})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error deleting region: %s", err))
		return
	}

	// wait until region is gone
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		_, err := p.client.GetRegion(ctx, model.GetName())
		if err != nil {
			if diagrid_errors.IsResourceNotFoundError(err) {
				return true, nil
			}

			return false, fmt.Errorf("error checking for deleted region: %w", err)
		}

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting region: %s", err))
		return
	}

	tflog.Debug(ctx, "deleted region",
		map[string]interface{}{
			"name": model.GetName(),
		})
}

func (p *regionResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	model := NewModel()
	model.SetName(req.ID)

	if err := read(ctx, p.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "region not found",
				map[string]interface{}{
					"name": model.GetName(),
				})

			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported region: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
