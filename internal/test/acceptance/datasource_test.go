package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccOrganizationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.catalyst_organization.test", "id"),
					resource.TestCheckResourceAttrSet("data.catalyst_organization.test", "name"),
				),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig() string {
	return `
data "catalyst_organization" "test" {}
`
}

func TestAccProjectDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	projectName := fmt.Sprintf("tfacc-prj-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a project and read it back with data source
			{
				Config: testAccProjectDataSourceConfig(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
					resource.TestCheckResourceAttr("data.catalyst_project.test", "name", projectName),
					resource.TestCheckResourceAttrPair("catalyst_project.test", "grpc_endpoint", "data.catalyst_project.test", "grpc_endpoint"),
					resource.TestCheckResourceAttrPair("catalyst_project.test", "http_endpoint", "data.catalyst_project.test", "http_endpoint"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(projectName string) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
  name = %[1]q
}

data "catalyst_project" "test" {
  name = catalyst_project.test.name
}
`, projectName)
}

func TestAccAppIdDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	appidName := fmt.Sprintf("tfacc-aid-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create an appid and read it back with data source
			{
				Config: testAccAppIdDataSourceConfig(projectID, appidName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_appid.test", "name", appidName),
					resource.TestCheckResourceAttr("data.catalyst_appid.test", "name", appidName),
					resource.TestCheckResourceAttrPair("catalyst_appid.test", "project_id", "data.catalyst_appid.test", "project_id"),
					resource.TestCheckResourceAttrPair("catalyst_appid.test", "app_config", "data.catalyst_appid.test", "app_config"),
					resource.TestCheckResourceAttrPair("catalyst_appid.test", "protocol", "data.catalyst_appid.test", "protocol"),
				),
			},
		},
	})
}

func testAccAppIdDataSourceConfig(projectID, appidName string) string {
	return fmt.Sprintf(`
resource "catalyst_appid" "test" {
  project_id = %[1]q
  name       = %[2]q
  protocol   = "http"
}

data "catalyst_appid" "test" {
  project_id = catalyst_appid.test.project_id
  name       = catalyst_appid.test.name
}
`, projectID, appidName)
}

func TestAccComponentDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	componentName := fmt.Sprintf("tfacc-cmp-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a component and read it back with data source
			{
				Config: testAccComponentDataSourceConfig(projectID, componentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
					resource.TestCheckResourceAttr("data.catalyst_component.test", "name", componentName),
					resource.TestCheckResourceAttrPair("catalyst_component.test", "type", "data.catalyst_component.test", "type"),
					resource.TestCheckResourceAttrPair("catalyst_component.test", "version", "data.catalyst_component.test", "version"),
					resource.TestCheckResourceAttrPair("catalyst_component.test", "status", "data.catalyst_component.test", "status"),
				),
			},
		},
	})
}

func testAccComponentDataSourceConfig(projectID, componentName string) string {
	return fmt.Sprintf(`
resource "catalyst_component" "test" {
  project_name = %[1]q
  name         = %[2]q
  type         = "state.redis"
  version      = "v1"
  spec         = <<-EOT
    - name: redisHost
      value: localhost:6379
    - name: redisPassword
      value: testpass123
  EOT
}

data "catalyst_component" "test" {
  project_name = catalyst_component.test.project_name
  name         = catalyst_component.test.name
}
`, projectID, componentName)
}
