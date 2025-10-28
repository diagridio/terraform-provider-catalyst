resource "catalyst_project" "example" {
  name   = "example-project"
  region = "us-east-1"
}

resource "catalyst_pubsub" "example" {
  project_name     = catalyst_project.example.name
  name             = "my-pubsub"
  component_name   = "my-pubsub-component"
  create_component = true
}
