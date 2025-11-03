package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Spec      *specModel   `tfsdk:"spec"`
	Status    types.String `tfsdk:"status"`
}

type specModel struct {
	AccessControl   *accessControlModel `tfsdk:"access_control"`
	AppHTTPPipeline *pipelineModel      `tfsdk:"app_http_pipeline"`
	HttpPipeline    *pipelineModel      `tfsdk:"http_pipeline"`
}

type accessControlModel struct {
	DefaultAction types.String          `tfsdk:"default_action"`
	TrustDomain   types.String          `tfsdk:"trust_domain"`
	Policies      []accessControlPolicy `tfsdk:"policies"`
}

type accessControlPolicy struct {
	AppID         types.String             `tfsdk:"app_id"`
	DefaultAction types.String             `tfsdk:"default_action"`
	Namespace     types.String             `tfsdk:"namespace"`
	TrustDomain   types.String             `tfsdk:"trust_domain"`
	Operations    []accessControlOperation `tfsdk:"operations"`
}

type accessControlOperation struct {
	Name      types.String `tfsdk:"name"`
	Action    types.String `tfsdk:"action"`
	HTTPVerbs types.List   `tfsdk:"http_verbs"`
}

type pipelineModel struct {
	Handlers []handlerModel `tfsdk:"handlers"`
}

type handlerModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_id": m.GetProjectID(),
		"name":       m.GetName(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_id: %s, name: %s`,
		m.GetProjectID(),
		m.GetName())
}

func (m *model) GetProjectID() string {
	return m.ProjectID.ValueString()
}

func (m *model) SetProjectID(projectID string) {
	m.ProjectID = types.StringValue(projectID)
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetStatus() string {
	return m.Status.ValueString()
}

func (m *model) SetStatus(status string) {
	if status == "" {
		m.Status = types.StringNull()
		return
	}
	m.Status = types.StringValue(status)
}

func (m *model) ensureSpec() *specModel {
	if m.Spec == nil {
		m.Spec = &specModel{}
	}
	return m.Spec
}
