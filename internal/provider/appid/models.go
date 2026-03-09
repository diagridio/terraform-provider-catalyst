package appid

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type appEndpointModel struct {
	URL                  types.String `tfsdk:"url"`
	Token                types.String `tfsdk:"token"`
	TokenHeader          types.String `tfsdk:"token_header"`
	ClientTimeoutSeconds types.Int64  `tfsdk:"client_timeout_seconds"`
}

type healthCheckModel struct {
	Path             types.String `tfsdk:"path"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	FailureThreshold types.Int64  `tfsdk:"failure_threshold"`
	IntervalSeconds  types.Int64  `tfsdk:"interval_seconds"`
	TimeoutMs        types.Int64  `tfsdk:"timeout_ms"`
}

type model struct {
	ProjectID        types.String `tfsdk:"project_id"`
	Name             types.String `tfsdk:"name"`
	AppConfig        types.String `tfsdk:"app_config"`
	Protocol         types.String `tfsdk:"protocol"`
	ApiTokenRevision types.Int64  `tfsdk:"api_token_revision"`
	Status           types.String `tfsdk:"status"`
	AppEndpoint      types.Object `tfsdk:"app_endpoint"`
	HealthCheck      types.Object `tfsdk:"health_check"`
	MaxBodySize      types.String `tfsdk:"max_body_size"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_id":    m.GetProjectID(),
		"name":          m.GetName(),
		"max_body_size": m.MaxBodySize.ValueString(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_id: %s, name: %s, max_body_size: %s`,
		m.GetProjectID(),
		m.GetName(),
		m.MaxBodySize.ValueString())
}

func (m *model) GetMaxBodySize() *string {
	if m.MaxBodySize.IsNull() || m.MaxBodySize.IsUnknown() {
		return nil
	}
	v := m.MaxBodySize.ValueString()
	return &v
}

func (m *model) SetMaxBodySize(v *string) {
	if v == nil {
		m.MaxBodySize = types.StringNull()
	} else {
		m.MaxBodySize = types.StringValue(*v)
	}
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

func (m *model) GetAppConfig() string {
	return m.AppConfig.ValueString()
}

func (m *model) SetAppConfig(appConfig string) {
	if appConfig == "" {
		m.AppConfig = types.StringNull()
	} else {
		m.AppConfig = types.StringValue(appConfig)
	}
}

func (m *model) GetProtocol() string {
	return m.Protocol.ValueString()
}

func (m *model) SetProtocol(protocol string) {
	if protocol == "" {
		m.Protocol = types.StringNull()
	} else {
		m.Protocol = types.StringValue(protocol)
	}
}

func (m *model) GetApiTokenRevision() int64 {
	return m.ApiTokenRevision.ValueInt64()
}

func (m *model) SetApiTokenRevision(revision int) {
	if revision == 0 {
		m.ApiTokenRevision = types.Int64Null()
	} else {
		m.ApiTokenRevision = types.Int64Value(int64(revision))
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
