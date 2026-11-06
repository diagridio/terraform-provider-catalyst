output "kvstore_name" {
  description = "The name of the KVStore"
  value       = data.diagrid_kvstore.example.name
}

output "kvstore_component_name" {
  description = "The component name of the KVStore"
  value       = data.diagrid_kvstore.example.component_name
}

output "kvstore_create_component" {
  description = "Whether to create the component"
  value       = data.diagrid_kvstore.example.create_component
}
