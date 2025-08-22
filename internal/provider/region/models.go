package region

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// model describes the data source data model.
type model struct {
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Ingress   types.String `tfsdk:"ingress"`
	Location  types.String `tfsdk:"location"`
	Type      types.String `tfsdk:"type"`
	JoinToken types.String `tfsdk:"join_token"`
	Connected types.Bool   `tfsdk:"connected"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetHost() string {
	return m.Host.ValueString()
}

func (m *model) SetHost(host string) {
	m.Host = types.StringValue(host)
}

func (m *model) GetIngress() string {
	return m.Ingress.ValueString()
}

func (m *model) SetIngress(ingress string) {
	m.Ingress = types.StringValue(ingress)
}

func (m *model) GetLocation() string {
	return m.Location.ValueString()
}

func (m *model) SetLocation(location string) {
	m.Location = types.StringValue(location)
}

func (m *model) GetType() string {
	return m.Type.ValueString()
}

func (m *model) SetType(regionType string) {
	m.Type = types.StringValue(regionType)
}

func (m *model) GetJoinToken() string {
	return m.JoinToken.ValueString()
}

func (m *model) SetJoinToken(joinToken string) {
	m.JoinToken = types.StringValue(joinToken)
}

func (m *model) GetConnected() bool {
	return m.Connected.ValueBool()
}

func (m *model) SetConnected(connected bool) {
	m.Connected = types.BoolValue(connected)
}

func (m *model) String() string {
	return fmt.Sprintf(`name: %s,
	host: %s,
	ingress: %s,
	location: %s,
	type: %s,
	connected?: %t`,
		m.GetName(),
		m.GetHost(),
		m.GetIngress(),
		m.GetLocation(),
		m.GetType(),
		m.GetConnected(),
	)
}
