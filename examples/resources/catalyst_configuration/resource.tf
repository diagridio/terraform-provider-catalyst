resource "catalyst_project" "example" {
  name   = "example-project"
}

resource "catalyst_configuration" "example" {
  project_id = catalyst_project.example.name
  name       = "my-configuration"

  # Spec defines Dapr configuration as YAML
  spec = <<-EOT
    accessControl:
      defaultAction: deny
      policies:
      - appId: app1
        defaultAction: allow
        namespace: default
      - appId: app2
        defaultAction: deny
        namespace: default
    api:
      allowed:
      - name: "GET /healthz"
        protocol: "http"
    tracing:
      samplingRate: "1"
      zipkin:
        endpointAddress: "http://zipkin.default.svc.cluster.local:9411/api/v2/spans"
    metrics:
      enabled: true
    secrets:
      scopes:
      - storeName: "local-secret-store"
        defaultAccess: "allow"
  EOT
}
