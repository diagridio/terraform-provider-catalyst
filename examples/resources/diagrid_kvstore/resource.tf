resource "diagrid_project" "example" {
  name = "example-project"
}

resource "diagrid_kvstore" "example" {
  project_name     = diagrid_project.example.name
  name             = "my-kvstore"
  component_name   = "my-kvstore-component"
  create_component = true
}
