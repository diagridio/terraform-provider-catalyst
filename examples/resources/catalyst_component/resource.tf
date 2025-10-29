resource "catalyst_project" "example" {
  name   = "example-project"
}

resource "catalyst_component" "example" {
  project_name = catalyst_project.example.name
  name         = "my-component"
  type         = "state.redis"
  version      = "v1"

  # Metadata as YAML
  spec = <<-EOT
    - name: redisHost
      value: "redis:6379"
    - name: redisPassword
      secretKeyRef:
        name: redis-secret
        key: password
  EOT

  secret_store = "kubernetes"
  scopes       = ["app1", "app2"]
}
