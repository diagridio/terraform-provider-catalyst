resource "diagrid_project" "example" {
  name = "example-project"
}

resource "diagrid_pubsub" "example" {
  project_name     = diagrid_project.example.name
  name             = "my-pubsub"
  component_name   = "my-pubsub-component"
  create_component = true
}
