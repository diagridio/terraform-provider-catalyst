package component

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type secretKeyRefModel struct {
	Key  types.String `tfsdk:"key"`
	Name types.String `tfsdk:"name"`
}

type metadataItemModel struct {
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	SecretKeyRef types.Object `tfsdk:"secret_key_ref"`
}

type model struct {
	ProjectName types.String `tfsdk:"project_name"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Version     types.String `tfsdk:"version"`
	Spec        types.String `tfsdk:"spec"`
	SecretStore types.String `tfsdk:"secret_store"`
	Scopes      types.List   `tfsdk:"scopes"`
	Status      types.String `tfsdk:"status"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_name": m.GetProjectName(),
		"name":         m.GetName(),
		"type":         m.GetType(),
		"version":      m.GetVersion(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_name: %s, name: %s, type: %s, version: %s`,
		m.GetProjectName(),
		m.GetName(),
		m.GetType(),
		m.GetVersion())
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

func (m *model) GetType() string {
	return m.Type.ValueString()
}

func (m *model) SetType(t string) {
	m.Type = types.StringValue(t)
}

func (m *model) GetVersion() string {
	return m.Version.ValueString()
}

func (m *model) SetVersion(version string) {
	m.Version = types.StringValue(version)
}

func (m *model) GetSpec() types.String {
	return m.Spec
}

func (m *model) SetSpec(spec string) {
	m.Spec = types.StringValue(spec)
}

func (m *model) GetSecretStore() string {
	return m.SecretStore.ValueString()
}

func (m *model) SetSecretStore(secretStore string) {
	if secretStore == "" {
		m.SecretStore = types.StringNull()
	} else {
		m.SecretStore = types.StringValue(secretStore)
	}
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
