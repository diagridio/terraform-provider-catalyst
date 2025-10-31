output "resiliency_name" {
  description = "The name of the resiliency policy"
  value       = catalyst_resiliency.example.name
}

output "resiliency_project_id" {
  description = "The project ID of the resiliency policy"
  value       = catalyst_resiliency.example.project_id
}
