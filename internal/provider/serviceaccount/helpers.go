package serviceaccount

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading service account",
		map[string]interface{}{
			"name": m.GetName(),
		})

	serviceAccount, err := client.GetServiceAccount(ctx, m.GetName())
	if err != nil {
		return fmt.Errorf("error getting service account: %w", err)
	}

	tflog.Debug(ctx, "read service account",
		map[string]interface{}{
			"serviceAccount": serviceAccount,
		})

	if serviceAccount.Metadata != nil {
		if serviceAccount.Metadata.Name != nil {
			m.SetName(*serviceAccount.Metadata.Name)
		}
	}

	m.SetDescription(serviceAccount.Spec.Description)
	m.SetOwner(serviceAccount.Spec.Owner)
	m.SetRole(serviceAccount.Spec.Role)

	if serviceAccount.Status != nil {
		if serviceAccount.Status.Email != nil {
			m.SetEmail(*serviceAccount.Status.Email)
		}
	}

	m.Log(ctx, "read service account model")

	return nil
}
