resource "catalyst_project" "example" {
  name   = "example-project"
}

resource "catalyst_appid" "example" {
  project_id = catalyst_project.example.name
  name       = "my-app"
}
