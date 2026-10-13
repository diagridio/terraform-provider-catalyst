package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	Name                              types.String `tfsdk:"name"`
	Region                            types.String `tfsdk:"region"`
	GRPCEndpoint                      types.String `tfsdk:"grpc_endpoint"`
	HTTPEndpoint                      types.String `tfsdk:"http_endpoint"`
	DefaultAgentInfrastructureEnabled types.Bool   `tfsdk:"default_agent_infrastructure_enabled"`
	DefaultKVStoreEnabled             types.Bool   `tfsdk:"default_kvstore_enabled"`
	DefaultPubsubEnabled              types.Bool   `tfsdk:"default_pubsub_enabled"`
	DefaultWorkflowStoreEnabled       types.Bool   `tfsdk:"default_workflow_store_enabled"`
	DisableAppTunnels                 types.Bool   `tfsdk:"disable_app_tunnels"`
	PrivateRegion                     types.Bool   `tfsdk:"private_region"`
	GlobalAppIdMaxBodySize            types.String `tfsdk:"global_app_id_max_body_size"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"name":                                 m.GetName(),
		"region":                               m.GetRegion(),
		"grpc_endpoint":                        m.GRPCEndpoint.ValueString(),
		"http_endpoint":                        m.HTTPEndpoint.ValueString(),
		"default_agent_infrastructure_enabled": m.DefaultAgentInfrastructureEnabled.ValueBool(),
		"default_kvstore_enabled":              m.DefaultKVStoreEnabled.ValueBool(),
		"default_pubsub_enabled":               m.DefaultPubsubEnabled.ValueBool(),
		"default_workflow_store_enabled":       m.DefaultWorkflowStoreEnabled.ValueBool(),
		"disable_app_tunnels":                  m.DisableAppTunnels.ValueBool(),
		"private_region":                       m.PrivateRegion.ValueBool(),
		"global_app_id_max_body_size":          m.GlobalAppIdMaxBodySize.ValueString(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`name: %s,
		region: %s,
		grpc_endpoint: %s,
		http_endpoint: %s,
		default_agent_infrastructure_enabled: %v,
		default_kvstore_enabled: %v,
		default_pubsub_enabled: %v,
		default_workflow_store_enabled: %v,
		disable_app_tunnels: %v,
		private_region: %v,
		global_app_id_max_body_size: %s`,
		m.GetName(),
		m.GetRegion(),
		m.GetGRPCEndpoint(),
		m.GetHTTPEndpoint(),
		m.DefaultAgentInfrastructureEnabled.ValueBool(),
		m.DefaultKVStoreEnabled.ValueBool(),
		m.DefaultPubsubEnabled.ValueBool(),
		m.DefaultWorkflowStoreEnabled.ValueBool(),
		m.DisableAppTunnels.ValueBool(),
		m.PrivateRegion.ValueBool(),
		m.GlobalAppIdMaxBodySize.ValueString())
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetRegion() string {
	return m.Region.ValueString()
}

func (m *model) SetRegion(region string) {
	m.Region = types.StringValue(region)
}

func (m *model) GetGRPCEndpoint() string {
	return m.GRPCEndpoint.ValueString()
}

func (m *model) SetGRPCEndpoint(endpoint string) {
	m.GRPCEndpoint = types.StringValue(endpoint)
}

func (m *model) GetHTTPEndpoint() string {
	return m.HTTPEndpoint.ValueString()
}

func (m *model) SetHTTPEndpoint(endpoint string) {
	m.HTTPEndpoint = types.StringValue(endpoint)
}

func (m *model) GetDefaultAgentInfrastructureEnabled() *bool {
	if m.DefaultAgentInfrastructureEnabled.IsNull() || m.DefaultAgentInfrastructureEnabled.IsUnknown() {
		return nil
	}
	v := m.DefaultAgentInfrastructureEnabled.ValueBool()
	return &v
}

func (m *model) SetDefaultAgentInfrastructureEnabled(v *bool) {
	if v == nil {
		m.DefaultAgentInfrastructureEnabled = types.BoolNull()
	} else {
		m.DefaultAgentInfrastructureEnabled = types.BoolValue(*v)
	}
}

func (m *model) GetDefaultKVStoreEnabled() *bool {
	if m.DefaultKVStoreEnabled.IsNull() || m.DefaultKVStoreEnabled.IsUnknown() {
		return nil
	}
	v := m.DefaultKVStoreEnabled.ValueBool()
	return &v
}

func (m *model) SetDefaultKVStoreEnabled(v *bool) {
	if v == nil {
		m.DefaultKVStoreEnabled = types.BoolNull()
	} else {
		m.DefaultKVStoreEnabled = types.BoolValue(*v)
	}
}

func (m *model) GetDefaultPubsubEnabled() *bool {
	if m.DefaultPubsubEnabled.IsNull() || m.DefaultPubsubEnabled.IsUnknown() {
		return nil
	}
	v := m.DefaultPubsubEnabled.ValueBool()
	return &v
}

func (m *model) SetDefaultPubsubEnabled(v *bool) {
	if v == nil {
		m.DefaultPubsubEnabled = types.BoolNull()
	} else {
		m.DefaultPubsubEnabled = types.BoolValue(*v)
	}
}

func (m *model) GetDefaultWorkflowStoreEnabled() *bool {
	if m.DefaultWorkflowStoreEnabled.IsNull() || m.DefaultWorkflowStoreEnabled.IsUnknown() {
		return nil
	}
	v := m.DefaultWorkflowStoreEnabled.ValueBool()
	return &v
}

func (m *model) SetDefaultWorkflowStoreEnabled(v *bool) {
	if v == nil {
		m.DefaultWorkflowStoreEnabled = types.BoolNull()
	} else {
		m.DefaultWorkflowStoreEnabled = types.BoolValue(*v)
	}
}

func (m *model) GetDisableAppTunnels() *bool {
	if m.DisableAppTunnels.IsNull() || m.DisableAppTunnels.IsUnknown() {
		return nil
	}
	v := m.DisableAppTunnels.ValueBool()
	return &v
}

func (m *model) SetDisableAppTunnels(v *bool) {
	if v == nil {
		m.DisableAppTunnels = types.BoolNull()
	} else {
		m.DisableAppTunnels = types.BoolValue(*v)
	}
}

func (m *model) GetPrivateRegion() *bool {
	if m.PrivateRegion.IsNull() || m.PrivateRegion.IsUnknown() {
		return nil
	}
	v := m.PrivateRegion.ValueBool()
	return &v
}

func (m *model) SetPrivateRegion(v *bool) {
	if v == nil {
		m.PrivateRegion = types.BoolNull()
	} else {
		m.PrivateRegion = types.BoolValue(*v)
	}
}

func (m *model) GetGlobalAppIdMaxBodySize() *string {
	if m.GlobalAppIdMaxBodySize.IsNull() || m.GlobalAppIdMaxBodySize.IsUnknown() {
		return nil
	}
	v := m.GlobalAppIdMaxBodySize.ValueString()
	return &v
}

func (m *model) SetGlobalAppIdMaxBodySize(v *string) {
	if v == nil {
		m.GlobalAppIdMaxBodySize = types.StringNull()
	} else {
		m.GlobalAppIdMaxBodySize = types.StringValue(*v)
	}
}
