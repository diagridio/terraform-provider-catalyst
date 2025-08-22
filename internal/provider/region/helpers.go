package region

import (
	"context"
	"fmt"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading region",
		map[string]interface{}{
			"region": m.GetName(),
		})

	region, err := client.GetRegion(ctx, m.GetName())
	if err != nil {
		return fmt.Errorf("error getting region: %w", err)
	}

	tflog.Debug(ctx, "read region",
		map[string]interface{}{
			"name":       *region.Metadata.Name,
			"type":       *region.Spec.Type,
			"ingress":    *region.Spec.Ingress,
			"host":       *region.Spec.Host,
			"location":   *region.Spec.Location,
			"connected?": region.Status.Connected != nil && *region.Status.Connected,
		})

	m.SetName(*region.Metadata.Name)
	m.SetIngress(*region.Spec.Ingress)
	m.SetType(*region.Spec.Type)
	if region.Spec.Host != nil &&
		*region.Spec.Host != "" {
		m.SetHost(*region.Spec.Host)
	}
	if region.Spec.Location != nil &&
		*region.Spec.Location != "" {
		m.SetLocation(*region.Spec.Location)
	}
	m.SetConnected(false)
	if region.Status.Connected != nil {
		m.SetConnected(*region.Status.Connected)
	}

	return nil

}
