# Example showing how to access the sensitive token
output "api_key_token" {
  value       = data.catalyst_service_account_api_key.api_key.token
  sensitive   = true
  description = "API key token"
}
