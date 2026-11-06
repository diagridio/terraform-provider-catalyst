resource "diagrid_project" "example" {
  name = "example-project"
}

resource "diagrid_configuration" "example" {
  project_id = diagrid_project.example.name
  name       = "my-configuration"

  spec = {
    access_control = {
      default_action = "allow"
      policies = [
        {
          app_id         = "app1"
          default_action = "allow"
          operations = [
            {
              name       = "op1"
              action     = "allow"
              http_verbs = ["GET", "POST"]
            }
          ]
        }
      ]
    }

    app_http_pipeline = {
      handlers = [
        {
          name = "oauth2"
          type = "middleware.http.oauth2"
        }
      ]
    }

    http_pipeline = {
      handlers = [
        {
          name = "ratelimit"
          type = "middleware.http.ratelimit"
        }
      ]
    }
  }
}
