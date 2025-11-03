resource "catalyst_project" "example" {
  name = "example-project"
}

resource "catalyst_configuration" "example" {
  project_id = catalyst_project.example.name
  name       = "my-configuration"

  spec = {
    access_control = {
      default_action = "allow"
      trust_domain   = "public"
      policies = [
        {
          app_id         = "app1"
          default_action = "allow"
          # namespace is optional and will be computed by the API if not provided
          namespace = "default"
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
