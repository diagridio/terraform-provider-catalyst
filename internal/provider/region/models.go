package region

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// model describes the data source data model.
type model struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	CloudProvider       types.String `tfsdk:"cloud_provider"`
	CloudProviderRegion types.String `tfsdk:"cloud_provider_region"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) GetID() string {
	return m.ID.ValueString()
}

func (m *model) SetID(id string) {
	m.ID = types.StringValue(id)
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetCloudProvider() string {
	return m.CloudProvider.ValueString()
}

func (m *model) SetCloudProvider(provider string) {
	m.CloudProvider = types.StringValue(provider)
}

func (m *model) GetCloudProviderRegion() string {
	return m.CloudProviderRegion.ValueString()
}

func (m *model) SetCloudProviderRegion(region string) {
	m.CloudProviderRegion = types.StringValue(region)
}
