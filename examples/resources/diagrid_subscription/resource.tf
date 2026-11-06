resource "diagrid_project" "example" {
  name = "example-project"
}

resource "diagrid_pubsub" "example" {
  project_name     = diagrid_project.example.name
  name             = "my-pubsub"
  component_name   = "my-pubsub-component"
  create_component = true
}

resource "diagrid_subscription" "example" {
  project_name = diagrid_project.example.name
  name         = "my-subscription"
  pubsub_name  = diagrid_pubsub.example.name
  topic        = "my-topic"

  scopes = ["app1"]

  # Spec defines the Dapr subscription structure using nested attributes
  spec = {
    routes = {
      default = "/orders"
      rules = [
        {
          match = "event.type == \"premium\""
          path  = "/premium-orders"
        }
      ]
    }

    bulk_subscribe = {
      enabled               = true
      max_messages_count    = 100
      max_await_duration_ms = 1000
    }

    dead_letter_topic = "orders-dlq"

    metadata = {
      rawPayload = "true"
    }
  }
}
