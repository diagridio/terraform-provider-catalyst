resource "catalyst_project" "example" {
  name = "example-project"
}

resource "catalyst_pubsub" "example" {
  project_name     = catalyst_project.example.name
  name             = "my-pubsub"
  component_name   = "my-pubsub-component"
  create_component = true
}

resource "catalyst_subscription" "example" {
  project_name = catalyst_project.example.name
  name         = "my-subscription"
  pubsub_name  = catalyst_pubsub.example.name
  topic        = "my-topic"

  scopes = ["app1"]

  # Spec defines Dapr Subscription as YAML
  spec = <<-EOT
    routes:
      default: /orders
      rules:
      - match: event.type == "premium"
        path: /premium-orders
    bulkSubscribe:
      enabled: true
      maxMessagesCount: 100
      maxAwaitDurationMs: 1000
    deadLetterTopic: orders-dlq
  EOT
}
