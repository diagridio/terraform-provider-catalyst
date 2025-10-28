output "appid_name" {
  description = "The name of the AppID"
  value       = data.catalyst_appid.example.name
}

output "appid_project_id" {
  description = "The project ID of the AppID"
  value       = data.catalyst_appid.example.project_id
}
