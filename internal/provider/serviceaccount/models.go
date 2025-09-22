package serviceaccount

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Owner       types.String `tfsdk:"owner"`
	Role        types.String `tfsdk:"role"`
	Email       types.String `tfsdk:"email"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"name":        m.GetName(),
		"description": m.GetDescription(),
		"owner":       m.GetOwner(),
		"role":        m.GetRole(),
		"email":       m.GetEmail(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`name: %s,
		description: %s,
		owner: %s,
		role: %s,
		email: %s`,
		m.GetName(),
		m.GetDescription(),
		m.GetOwner(),
		m.GetRole(),
		m.GetEmail())
}

func (m *model) GetName() string {
	return m.Name.ValueString()
}

func (m *model) SetName(name string) {
	m.Name = types.StringValue(name)
}

func (m *model) GetDescription() string {
	return m.Description.ValueString()
}

func (m *model) SetDescription(description string) {
	m.Description = types.StringValue(description)
}

func (m *model) GetOwner() string {
	return m.Owner.ValueString()
}

func (m *model) SetOwner(owner string) {
	m.Owner = types.StringValue(owner)
}

func (m *model) GetRole() string {
	return m.Role.ValueString()
}

func (m *model) SetRole(role string) {
	m.Role = types.StringValue(role)
}

func (m *model) GetEmail() string {
	return m.Email.ValueString()
}

func (m *model) SetEmail(email string) {
	m.Email = types.StringValue(email)
}
