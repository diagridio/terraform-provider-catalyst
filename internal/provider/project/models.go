package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TODO: move this somewhere else
const (
	CatalystDiagridV1Beta1 = "cra.diagrid.io/v1beta1"
	Project                = "Project"
)

type model struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	Name            types.String `tfsdk:"name"`
	Region          types.String `tfsdk:"region"`
	GRPCEndpoint    types.String `tfsdk:"grpc_endpoint"`
	HTTPEndpoint    types.String `tfsdk:"http_endpoint"`
	ManagedPubsub   types.Bool   `tfsdk:"managed_pubsub"`
	ManagedKVStore  types.Bool   `tfsdk:"managed_kvstore"`
	ManagedWorkflow types.Bool   `tfsdk:"managed_workflow"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"organization_id":  m.GetOrganizationID(),
		"name":             m.GetName(),
		"region":           m.GetRegion(),
		"grpc_endpoint":    m.GRPCEndpoint.ValueString(),
		"http_endpoint":    m.HTTPEndpoint.ValueString(),
		"managed_pubsub":   m.GetManagedPubsub(),
		"managed_kvstore":  m.GetManagedKVStore(),
		"managed_workflow": m.GetManagedWorkflow(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`organization_id: %s,
		name: %s,
		region: %s,
		grpc_endpoint: %s,
		http_endpoint: %s,
		managed_pubsub: %t,
		managed_kvstore: %t,
		managed_workflow: %t`,
		m.GetOrganizationID(),
		m.GetName(),
		m.GetRegion(),
		m.GetGRPCEndpoint(),
		m.GetHTTPEndpoint(),
		m.GetManagedPubsub(),
		m.GetManagedKVStore(),
		m.GetManagedWorkflow())
}

func (m *model) GetOrganizationID() string {
	return m.OrganizationID.ValueString()
}

func (m *model) SetOrganizationID(organizationID string) {
	m.OrganizationID = types.StringValue(organizationID)
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

func (m *model) GetManagedPubsub() bool {
	return m.ManagedPubsub.ValueBool()
}

func (m *model) SetManagedPubsub(b bool) {
	m.ManagedPubsub = types.BoolValue(b)
}

func (m *model) GetManagedKVStore() bool {
	return m.ManagedKVStore.ValueBool()
}

func (m *model) SetManagedKVStore(b bool) {
	m.ManagedKVStore = types.BoolValue(b)
}

func (m *model) GetManagedWorkflow() bool {
	return m.ManagedWorkflow.ValueBool()
}

func (m *model) SetManagedWorkflow(b bool) {
	m.ManagedWorkflow = types.BoolValue(b)
}
