package resiliency

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Spec      *specModel   `tfsdk:"spec"`
	Scopes    types.List   `tfsdk:"scopes"`
	Status    types.String `tfsdk:"status"`
}

type specModel struct {
	Policies *policiesModel `tfsdk:"policies"`
	Targets  *targetsModel  `tfsdk:"targets"`
}

type policiesModel struct {
	Timeouts        types.Map                      `tfsdk:"timeouts"`
	Retries         map[string]retryPolicyModel    `tfsdk:"retries"`
	CircuitBreakers map[string]circuitBreakerModel `tfsdk:"circuit_breakers"`
}

type retryPolicyModel struct {
	Duration    types.String `tfsdk:"duration"`
	MaxInterval types.String `tfsdk:"max_interval"`
	MaxRetries  types.Int64  `tfsdk:"max_retries"`
	Policy      types.String `tfsdk:"policy"`
}

type circuitBreakerModel struct {
	Interval    types.String `tfsdk:"interval"`
	MaxRequests types.Int64  `tfsdk:"max_requests"`
	Timeout     types.String `tfsdk:"timeout"`
	Trip        types.String `tfsdk:"trip"`
}

type targetsModel struct {
	Apps       map[string]endpointPolicyModel  `tfsdk:"apps"`
	Actors     map[string]actorPolicyModel     `tfsdk:"actors"`
	Components map[string]componentPolicyModel `tfsdk:"components"`
}

type endpointPolicyModel struct {
	CircuitBreaker          types.String `tfsdk:"circuit_breaker"`
	CircuitBreakerCacheSize types.Int64  `tfsdk:"circuit_breaker_cache_size"`
	Retry                   types.String `tfsdk:"retry"`
	Timeout                 types.String `tfsdk:"timeout"`
}

type actorPolicyModel struct {
	CircuitBreaker          types.String `tfsdk:"circuit_breaker"`
	CircuitBreakerCacheSize types.Int64  `tfsdk:"circuit_breaker_cache_size"`
	CircuitBreakerScope     types.String `tfsdk:"circuit_breaker_scope"`
	Retry                   types.String `tfsdk:"retry"`
	Timeout                 types.String `tfsdk:"timeout"`
}

type componentPolicyModel struct {
	Inbound  *componentDirectionPolicyModel `tfsdk:"inbound"`
	Outbound *componentDirectionPolicyModel `tfsdk:"outbound"`
}

type componentDirectionPolicyModel struct {
	CircuitBreaker types.String `tfsdk:"circuit_breaker"`
	Retry          types.String `tfsdk:"retry"`
	Timeout        types.String `tfsdk:"timeout"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_id": m.GetProjectID(),
		"name":       m.GetName(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_id: %s, name: %s`,
		m.GetProjectID(),
		m.GetName())
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

func (m *model) GetStatus() string {
	return m.Status.ValueString()
}

func (m *model) SetStatus(status string) {
	if status == "" {
		m.Status = types.StringNull()
		return
	}
	m.Status = types.StringValue(status)
}

func (m *model) ensureSpec() *specModel {
	if m.Spec == nil {
		m.Spec = &specModel{}
	}
	return m.Spec
}
