output "subscription_name" {
  description = "The name of the subscription"
  value       = data.diagrid_subscription.example.name
}

output "subscription_pubsub_name" {
  description = "The PubSub name of the subscription"
  value       = data.diagrid_subscription.example.pubsub_name
}

output "subscription_topic" {
  description = "The topic of the subscription"
  value       = data.diagrid_subscription.example.topic
}
