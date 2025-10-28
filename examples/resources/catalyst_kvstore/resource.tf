resource "catalyst_project" "example" {
  name   = "example-project"
  region = "us-east-1"
}

resource "catalyst_kvstore" "example" {
  project_name     = catalyst_project.example.name
  name             = "my-kvstore"
  component_name   = "my-kvstore-component"
  create_component = true
}
