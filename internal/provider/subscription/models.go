package subscription

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type model struct {
	ProjectName types.String `tfsdk:"project_name"`
	Name        types.String `tfsdk:"name"`
	PubsubName  types.String `tfsdk:"pubsub_name"`
	Topic       types.String `tfsdk:"topic"`
	Spec        *specModel   `tfsdk:"spec"`
	Scopes      types.List   `tfsdk:"scopes"`
	Status      types.String `tfsdk:"status"`
}

type specModel struct {
	Routes          *routesModel        `tfsdk:"routes"`
	BulkSubscribe   *bulkSubscribeModel `tfsdk:"bulk_subscribe"`
	DeadLetterTopic types.String        `tfsdk:"dead_letter_topic"`
	Metadata        types.Map           `tfsdk:"metadata"`
	Dynamic         types.Bool          `tfsdk:"dynamic"`
}

type routesModel struct {
	Default types.String     `tfsdk:"default"`
	Rules   []routeRuleModel `tfsdk:"rules"`
}

type routeRuleModel struct {
	Match types.String `tfsdk:"match"`
	Path  types.String `tfsdk:"path"`
}

type bulkSubscribeModel struct {
	Enabled            types.Bool  `tfsdk:"enabled"`
	MaxMessagesCount   types.Int64 `tfsdk:"max_messages_count"`
	MaxAwaitDurationMs types.Int64 `tfsdk:"max_await_duration_ms"`
}

func NewModel() *model {
	return &model{}
}

func (m *model) Log(ctx context.Context, msg string) {
	tflog.Debug(ctx, msg, map[string]interface{}{
		"project_name": m.GetProjectName(),
		"name":         m.GetName(),
		"pubsub_name":  m.GetPubsubName(),
		"topic":        m.GetTopic(),
	})
}

func (m *model) String() string {
	return fmt.Sprintf(`project_name: %s, name: %s, pubsub_name: %s, topic: %s`,
		m.GetProjectName(),
		m.GetName(),
		m.GetPubsubName(),
		m.GetTopic())
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

func (m *model) GetPubsubName() string {
	return m.PubsubName.ValueString()
}

func (m *model) SetPubsubName(pubsubName string) {
	m.PubsubName = types.StringValue(pubsubName)
}

func (m *model) GetTopic() string {
	return m.Topic.ValueString()
}

func (m *model) SetTopic(topic string) {
	m.Topic = types.StringValue(topic)
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
