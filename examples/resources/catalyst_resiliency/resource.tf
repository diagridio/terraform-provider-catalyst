resource "catalyst_project" "example" {
  name = "example-project"
}

resource "catalyst_resiliency" "example" {
  project_id = catalyst_project.example.name
  name       = "my-resiliency-policy"

  # Scopes define which apps use this resiliency policy
  scopes = ["app1", "app2"]

  # Spec defines the Dapr resiliency structure using nested attributes
  spec = {
    policies = {
      retries = {
        DefaultRetryPolicy = {
          policy      = "constant"
          duration    = "5s"
          max_retries = 10
        }
      }

      timeouts = {
        DefaultTimeoutPolicy = "60s"
      }

      circuit_breakers = {
        DefaultCircuitBreakerPolicy = {
          max_requests = 1
          interval     = "30s"
          timeout      = "60s"
          trip         = "consecutiveFailures > 5"
        }
      }
    }

    targets = {
      apps = {
        app1 = {
          retry           = "DefaultRetryPolicy"
          timeout         = "DefaultTimeoutPolicy"
          circuit_breaker = "DefaultCircuitBreakerPolicy"
        }

        app2 = {
          retry           = "DefaultRetryPolicy"
          timeout         = "DefaultTimeoutPolicy"
          circuit_breaker = "DefaultCircuitBreakerPolicy"
        }
      }

      components = {
        statestore = {
          outbound = {
            retry   = "DefaultRetryPolicy"
            timeout = "DefaultTimeoutPolicy"
          }
        }
      }
    }
  }
}
