package component

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	ProjectName types.String `tfsdk:"project_name"`
	Name        types.String `tfsdk:"name"`
	Spec        *specModel   `tfsdk:"spec"`
	Auth        *authModel   `tfsdk:"auth"`
	Scopes      types.List   `tfsdk:"scopes"`
	Status      types.String `tfsdk:"status"`
}

type specModel struct {
	Type     types.String        `tfsdk:"type"`
	Version  types.String        `tfsdk:"version"`
	Metadata []metadataItemModel `tfsdk:"metadata"`
}

type metadataItemModel struct {
	Name         types.String       `tfsdk:"name"`
	Value        types.String       `tfsdk:"value"`
	SecretKeyRef *secretKeyRefModel `tfsdk:"secret_key_ref"`
}

type secretKeyRefModel struct {
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

type authModel struct {
	SecretStore types.String `tfsdk:"secret_store"`
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
	if m.Spec == nil {
		return ""
	}
	return m.Spec.Type.ValueString()
}

func (m *model) SetType(t string) {
	spec := m.ensureSpec()
	if t == "" {
		spec.Type = types.StringNull()
		return
	}
	spec.Type = types.StringValue(t)
}

func (m *model) GetVersion() string {
	if m.Spec == nil {
		return ""
	}
	return m.Spec.Version.ValueString()
}

func (m *model) SetVersion(version string) {
	spec := m.ensureSpec()
	if version == "" {
		spec.Version = types.StringNull()
		return
	}
	spec.Version = types.StringValue(version)
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

func (m *model) ensureSpec() *specModel {
	if m.Spec == nil {
		m.Spec = &specModel{}
	}
	return m.Spec
}
