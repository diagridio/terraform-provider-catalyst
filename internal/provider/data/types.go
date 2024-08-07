package data

import (
	"github.com/diagridio/diagrid-cloud-go/cloudruntime"
	"github.com/diagridio/diagrid-cloud-go/management"
)

type ProviderData struct {
	ManagementClient *management.ManagementClient
	CatalystClient   cloudruntime.CloudruntimeAPIClient
}
