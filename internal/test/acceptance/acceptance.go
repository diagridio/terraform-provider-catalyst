package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/diagridio/terraform-provider-diagrid/internal/provider"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	provider.ProviderName: providerserver.NewProtocol6WithError(
		provider.New("test"),
	),
}

// TestAccPreCheck - Check if the environment variables are set.
func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("DIAGRID_API_KEY"); v == "" {
		t.Fatal("DIAGRID_API_KEY must be set for acceptance tests")
	}
}
