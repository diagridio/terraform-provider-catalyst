output "pubsub_name" {
  description = "The name of the PubSub"
  value       = catalyst_pubsub.example.name
}

output "pubsub_component_name" {
  description = "The component name of the PubSub"
  value       = catalyst_pubsub.example.component_name
}

output "pubsub_create_component" {
  description = "Whether to create the component"
  value       = catalyst_pubsub.example.create_component
}
