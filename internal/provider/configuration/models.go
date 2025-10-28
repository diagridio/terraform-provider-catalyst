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
	Spec      types.String `tfsdk:"spec"`
	Status    types.String `tfsdk:"status"`
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

func (m *model) GetSpec() types.String {
	return m.Spec
}

func (m *model) SetSpec(spec string) {
	m.Spec = types.StringValue(spec)
}

func (m *model) GetStatus() types.String {
	return m.Status
}

func (m *model) SetStatus(status string) {
	m.Status = types.StringValue(status)
}
