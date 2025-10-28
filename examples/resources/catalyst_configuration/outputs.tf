output "configuration_name" {
  description = "The name of the configuration"
  value       = catalyst_configuration.example.name
}

output "configuration_project_id" {
  description = "The project ID of the configuration"
  value       = catalyst_configuration.example.project_id
}
