package appid

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading appid",
		map[string]interface{}{
			"project_id": m.GetProjectID(),
			"name":       m.GetName(),
		})

	appid, err := client.GetAppId(ctx, m.GetProjectID(), m.GetName(), &cloudruntime_client.DescribeAppIdentityParams{})
	if err != nil {
		return fmt.Errorf("error getting appid: %w", err)
	}

	tflog.Debug(ctx, "read appid",
		map[string]interface{}{
			"appid": appid,
		})

	if appid.Metadata != nil && appid.Metadata.Name != nil {
		m.SetName(*appid.Metadata.Name)
	}

	if appid.Spec != nil {
		if appid.Spec.AppConfig != nil {
			m.SetAppConfig(*appid.Spec.AppConfig)
		} else {
			m.SetAppConfig("")
		}

		if appid.Spec.Protocol != nil {
			m.SetProtocol(*appid.Spec.Protocol)
		} else {
			m.SetProtocol("")
		}

		if appid.Spec.ApiTokenRevision != nil {
			m.SetApiTokenRevision(*appid.Spec.ApiTokenRevision)
		} else {
			m.SetApiTokenRevision(0)
		}

		// Handle AppEndpoint
		if appid.Spec.AppEndpoint != nil {
			appEndpointAttrs := map[string]attr.Value{
				"url":                    types.StringNull(),
				"token":                  types.StringNull(),
				"token_header":           types.StringNull(),
				"client_timeout_seconds": types.Int64Null(),
			}

			if appid.Spec.AppEndpoint.Url != nil {
				appEndpointAttrs["url"] = types.StringValue(*appid.Spec.AppEndpoint.Url)
			}
			if appid.Spec.AppEndpoint.Token != nil {
				appEndpointAttrs["token"] = types.StringValue(*appid.Spec.AppEndpoint.Token)
			}
			if appid.Spec.AppEndpoint.TokenHeader != nil {
				appEndpointAttrs["token_header"] = types.StringValue(*appid.Spec.AppEndpoint.TokenHeader)
			}
			if appid.Spec.AppEndpoint.ClientTimeoutSeconds != nil {
				appEndpointAttrs["client_timeout_seconds"] = types.Int64Value(int64(*appid.Spec.AppEndpoint.ClientTimeoutSeconds))
			}

			m.AppEndpoint = types.ObjectValueMust(
				map[string]attr.Type{
					"url":                    types.StringType,
					"token":                  types.StringType,
					"token_header":           types.StringType,
					"client_timeout_seconds": types.Int64Type,
				},
				appEndpointAttrs,
			)
		} else {
			m.AppEndpoint = types.ObjectNull(map[string]attr.Type{
				"url":                    types.StringType,
				"token":                  types.StringType,
				"token_header":           types.StringType,
				"client_timeout_seconds": types.Int64Type,
			})
		}

		// Handle HealthCheck
		if appid.Spec.HealthCheck != nil {
			healthCheckAttrs := map[string]attr.Value{
				"path":              types.StringNull(),
				"enabled":           types.BoolNull(),
				"failure_threshold": types.Int64Null(),
				"interval_seconds":  types.Int64Null(),
				"timeout_ms":        types.Int64Null(),
			}

			if appid.Spec.HealthCheck.Path != nil {
				healthCheckAttrs["path"] = types.StringValue(*appid.Spec.HealthCheck.Path)
			}
			if appid.Spec.HealthCheck.Probe != nil {
				healthCheckAttrs["enabled"] = types.BoolValue(appid.Spec.HealthCheck.Probe.Enabled)
				if appid.Spec.HealthCheck.Probe.FailureThreshold != nil {
					healthCheckAttrs["failure_threshold"] = types.Int64Value(int64(*appid.Spec.HealthCheck.Probe.FailureThreshold))
				}
				if appid.Spec.HealthCheck.Probe.IntervalInSec != nil {
					healthCheckAttrs["interval_seconds"] = types.Int64Value(int64(*appid.Spec.HealthCheck.Probe.IntervalInSec))
				}
				if appid.Spec.HealthCheck.Probe.TimeoutInMs != nil {
					healthCheckAttrs["timeout_ms"] = types.Int64Value(int64(*appid.Spec.HealthCheck.Probe.TimeoutInMs))
				}
			}

			m.HealthCheck = types.ObjectValueMust(
				map[string]attr.Type{
					"path":              types.StringType,
					"enabled":           types.BoolType,
					"failure_threshold": types.Int64Type,
					"interval_seconds":  types.Int64Type,
					"timeout_ms":        types.Int64Type,
				},
				healthCheckAttrs,
			)
		} else {
			m.HealthCheck = types.ObjectNull(map[string]attr.Type{
				"path":              types.StringType,
				"enabled":           types.BoolType,
				"failure_threshold": types.Int64Type,
				"interval_seconds":  types.Int64Type,
				"timeout_ms":        types.Int64Type,
			})
		}
	}

	// Set status (read-only)
	if appid.Status != nil {
		m.SetStatus(appid.GetStatus())
	} else {
		m.SetStatus("")
	}

	m.Log(ctx, "read appid model")

	return nil
}

func toAPIAppEndpoint(ctx context.Context, obj types.Object) *cloudruntime_client.AppIdentitySpecAppEndpoint {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var model appEndpointModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	endpoint := &cloudruntime_client.AppIdentitySpecAppEndpoint{}

	if !model.URL.IsNull() && !model.URL.IsUnknown() {
		url := model.URL.ValueString()
		endpoint.Url = &url
	}

	if !model.Token.IsNull() && !model.Token.IsUnknown() {
		token := model.Token.ValueString()
		endpoint.Token = &token
	}

	if !model.TokenHeader.IsNull() && !model.TokenHeader.IsUnknown() {
		tokenHeader := model.TokenHeader.ValueString()
		endpoint.TokenHeader = &tokenHeader
	}

	if !model.ClientTimeoutSeconds.IsNull() && !model.ClientTimeoutSeconds.IsUnknown() {
		timeout := int(model.ClientTimeoutSeconds.ValueInt64())
		endpoint.ClientTimeoutSeconds = &timeout
	}

	return endpoint
}

func toAPIHealthCheck(ctx context.Context, obj types.Object) *cloudruntime_client.AppIdentitySpecAppEndpointHealthCheck {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var model healthCheckModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	healthCheck := &cloudruntime_client.AppIdentitySpecAppEndpointHealthCheck{}

	if !model.Path.IsNull() && !model.Path.IsUnknown() {
		path := model.Path.ValueString()
		healthCheck.Path = &path
	}

	// Set probe if any probe fields are set
	if !model.Enabled.IsNull() || !model.FailureThreshold.IsNull() || !model.IntervalSeconds.IsNull() || !model.TimeoutMs.IsNull() {
		probe := &cloudruntime_client.AppIdentitySpecAppEndpointHealthCheckProbe{}

		if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
			probe.Enabled = model.Enabled.ValueBool()
		}

		if !model.FailureThreshold.IsNull() && !model.FailureThreshold.IsUnknown() {
			threshold := int32(model.FailureThreshold.ValueInt64())
			probe.FailureThreshold = &threshold
		}

		if !model.IntervalSeconds.IsNull() && !model.IntervalSeconds.IsUnknown() {
			interval := int32(model.IntervalSeconds.ValueInt64())
			probe.IntervalInSec = &interval
		}

		if !model.TimeoutMs.IsNull() && !model.TimeoutMs.IsUnknown() {
			timeout := int32(model.TimeoutMs.ValueInt64())
			probe.TimeoutInMs = &timeout
		}

		healthCheck.Probe = probe
	}

	return healthCheck
}
