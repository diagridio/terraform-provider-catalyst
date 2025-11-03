resource "catalyst_project" "example" {
  name = "example-project"
}

resource "catalyst_component" "example" {
  project_name = catalyst_project.example.name
  name         = "my-component"
  spec = {
    type    = "state.redis"
    version = "v1"

    metadata = [
      {
        name  = "redisHost"
        value = "redis:6379"
      },
      {
        name = "redisPassword"
        secret_key_ref = {
          name = "redis-secret"
          key  = "password"
        }
      }
    ]
  }

  auth = {
    secret_store = "kubernetes"
  }

  scopes       = ["app1", "app2"]
}
