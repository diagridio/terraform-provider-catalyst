package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectResource{}
var _ resource.ResourceWithImportState = &projectResource{}

// projectResource defines the resource implementation.
type projectResource struct {
	client catalyst.Client
}

func NewResource() resource.Resource {
	return &projectResource{}
}

func (p *projectResource) Metadata(ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (p *projectResource) Schema(ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Catalyst project resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Project region",
				Optional:            true,
			},
			"grpc_endpoint": schema.StringAttribute{
				MarkdownDescription: "gRPC endpoint",
				Optional:            true,
				Computed:            true,
			},
			"http_endpoint": schema.StringAttribute{
				MarkdownDescription: "HTTP endpoint",
				Optional:            true,
				Computed:            true,
			},
			"wait_for_ready": schema.BoolAttribute{
				MarkdownDescription: "Wait for the project to be in ready state before returning",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (p *projectResource) Configure(ctx context.Context,
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

func (p *projectResource) Create(ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	model := NewModel()

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "creating project",
		map[string]interface{}{
			"name":   model.GetName(),
			"region": model.GetRegion(),
		})

	project := &client.Project{
		ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
		Kind:       lo.ToPtr(catalyst.KindProject),
		Metadata: &client.Metadata{
			Name: lo.ToPtr(model.GetName()),
		},
		Spec: &client.ProjectSpec{
			DisplayName: lo.ToPtr(model.GetName()),
			Region:      lo.ToPtr(model.GetRegion()),
		},
		Status: &client.ProjectStatus{},
	}
	if err := p.client.CreateProject(ctx, project); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error creating project: %s", err))
		return
	}

	// wait until project is created and in ready status
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		project, err := p.client.GetProject(ctx, model.GetName(), &client.DescribeProjectParams{})
		if err != nil {
			return false, fmt.Errorf("Error getting project: %w", err)
		}

		expectedStatus := "processing"
		if model.WaitForReady.ValueBool() {
			expectedStatus = "ready"
		}

		if project.Status != nil &&
			project.Status.Status != nil &&
			*project.Status.Status == expectedStatus {
			return true, nil
		}

		tflog.Debug(ctx, "project status still not at expected value",
			map[string]interface{}{
				"name":     model.GetName(),
				"expected": expectedStatus,
			})

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting project: %s", err))
		return
	}

	tflog.Debug(ctx, "created project", map[string]interface{}{
		"id":     project.Metadata.Uid,
		"name":   *project.Metadata.Name,
		"region": *project.Spec.Region,
	})

	if err := read(ctx, p.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading created project: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (p *projectResource) Read(ctx context.Context,
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
			tflog.Debug(ctx, "project not found", map[string]interface{}{
				"name": model.GetName(),
			})

			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading project: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (p *projectResource) Update(ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	model := NewModel()

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := p.client.GetProject(ctx, model.GetName(), &client.DescribeProjectParams{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting project: %s", err))
		return
	}

	model.Log(ctx, "read project")

	project.Spec.DisplayName = lo.ToPtr(model.GetName())
	project.Spec.Region = lo.ToPtr(model.GetRegion())
	project.Status = &client.ProjectStatus{}

	tflog.Debug(ctx, "updating project", map[string]interface{}{
		"model": model.String(),
	})

	if err := p.client.UpdateProject(ctx, project); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error updating project: %s", err))
		return
	}

	// wait until project is created and in ready status
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		project, err := p.client.GetProject(ctx, model.GetName(), &client.DescribeProjectParams{})
		if err != nil {
			return false, fmt.Errorf("Error getting project: %w", err)
		}

		expectedStatus := "processing"
		if model.WaitForReady.ValueBool() {
			expectedStatus = "ready"
		}

		if project.Status != nil &&
			project.Status.Status != nil &&
			*project.Status.Status == expectedStatus {
			return true, nil
		}

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting project: %s", err))
		return
	}

	if err := read(ctx, p.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading updated project: %s", err))
		return
	}

	tflog.Debug(ctx, "updated project", map[string]interface{}{
		"model": model.String(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (p *projectResource) Delete(ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	model := NewModel()

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting project",
		map[string]interface{}{
			"name": model.GetName(),
		})

	if err := p.client.DeleteProject(ctx, model.GetName()); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "project to delete not found", map[string]interface{}{
				"name": model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error deleting project: %s", err))
		return
	}

	// wait until project is gone
	if err := helpers.WaitUntil(ctx, func(ctx context.Context) (bool, error) {
		_, err := p.client.GetProject(ctx, model.GetName(), &client.DescribeProjectParams{})
		if err != nil {
			if diagrid_errors.IsResourceNotFoundError(err) {
				return true, nil
			}

			return false, fmt.Errorf("error checking for deleted project: %w", err)
		}

		return false, nil
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Error getting project: %s", err))
		return
	}

	tflog.Debug(ctx, "deleted project",
		map[string]interface{}{
			"name": model.GetName(),
		})
}

func (p *projectResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	model := NewModel()
	model.SetName(req.ID)

	if err := read(ctx, p.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "project not found", map[string]interface{}{
				"name": model.GetName(),
			})

			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading imported project: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
