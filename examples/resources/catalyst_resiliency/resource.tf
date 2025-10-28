resource "catalyst_project" "example" {
  name   = "example-project"
  region = "us-east-1"
}

resource "catalyst_resiliency" "example" {
  project_id = catalyst_project.example.name
  name       = "my-resiliency-policy"
  
  # Scopes define which apps use this resiliency policy
  scopes = ["app1", "app2"]
  
  # Spec defines Dapr Resiliency as YAML
  spec = <<-EOT
    policies:
      retries:
        DefaultRetryPolicy:
          policy: constant
          duration: 5s
          maxRetries: 10
      timeouts:
        DefaultTimeoutPolicy: 60s
      circuitBreakers:
        DefaultCircuitBreakerPolicy:
          maxRequests: 1
          interval: 30s
          timeout: 60s
          trip: consecutiveFailures > 5
    targets:
      apps:
        app1:
          retry: DefaultRetryPolicy
          timeout: DefaultTimeoutPolicy
          circuitBreaker: DefaultCircuitBreakerPolicy
      components:
        statestore:
          outbound:
            retry: DefaultRetryPolicy
            timeout: DefaultTimeoutPolicy
  EOT
}
