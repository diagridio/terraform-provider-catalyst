package pubsub

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	ProjectName     types.String `tfsdk:"project_name"`
	Name            types.String `tfsdk:"name"`
	ComponentName   types.String `tfsdk:"component_name"`
	CreateComponent types.Bool   `tfsdk:"create_component"`
	Scopes          types.List   `tfsdk:"scopes"`
	Status          types.String `tfsdk:"status"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_name":     m.GetProjectName(),
		"name":             m.GetName(),
		"component_name":   m.GetComponentName(),
		"create_component": m.GetCreateComponent(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_name: %s, name: %s, component_name: %s, create_component: %t`,
		m.GetProjectName(),
		m.GetName(),
		m.GetComponentName(),
		m.GetCreateComponent())
}

func (m *model) GetProjectName() string {
	return m.ProjectName.ValueString()
}

func (m *model) SetProjectName(projectName string) {
	m.ProjectName = types.StringValue(projectName)
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetComponentName() string {
	return m.ComponentName.ValueString()
}

func (m *model) SetComponentName(componentName string) {
	m.ComponentName = types.StringValue(componentName)
}

func (m *model) GetCreateComponent() bool {
	return m.CreateComponent.ValueBool()
}

func (m *model) SetCreateComponent(createComponent bool) {
	m.CreateComponent = types.BoolValue(createComponent)
}

func (m *model) GetStatus() string {
	return m.Status.ValueString()
}

func (m *model) SetStatus(status string) {
	if status == "" {
		m.Status = types.StringNull()
	} else {
		m.Status = types.StringValue(status)
	}
}
