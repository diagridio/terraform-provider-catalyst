package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	Name         types.String `tfsdk:"name"`
	Region       types.String `tfsdk:"region"`
	GRPCEndpoint types.String `tfsdk:"grpc_endpoint"`
	HTTPEndpoint types.String `tfsdk:"http_endpoint"`
	WaitForReady types.Bool   `tfsdk:"wait_for_ready"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"name":          m.GetName(),
		"region":        m.GetRegion(),
		"grpc_endpoint": m.GRPCEndpoint.ValueString(),
		"http_endpoint": m.HTTPEndpoint.ValueString(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`name: %s,
		region: %s,
		grpc_endpoint: %s,
		http_endpoint: %s`,
		m.GetName(),
		m.GetRegion(),
		m.GetGRPCEndpoint(),
		m.GetHTTPEndpoint())
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
