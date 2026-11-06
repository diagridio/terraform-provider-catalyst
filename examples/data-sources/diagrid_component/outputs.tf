output "component_name" {
  description = "The name of the component"
  value       = data.diagrid_component.example.name
}

output "component_type" {
  description = "The type of the component"
  value       = data.diagrid_component.example.type
}

output "component_version" {
  description = "The version of the component"
  value       = data.diagrid_component.example.version
}
