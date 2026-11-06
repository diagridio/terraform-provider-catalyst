resource "diagrid_project" "example" {
  name = "example-project"
}

resource "diagrid_appid" "example" {
  project_id = diagrid_project.example.name
  name       = "my-app"
}
