package organization

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// model describes the data source data model.
type model struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Plan types.String `tfsdk:"plan"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) GetID() string {
	return m.ID.String()
}

func (m *model) SetID(id string) {
	m.ID = types.StringValue(id)
}

func (m *model) GetName() string {
	return m.Name.String()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetPlan() string {
	return m.Plan.String()
}

func (m *model) SetPlan(plan string) {
	m.Plan = types.StringValue(plan)
}
