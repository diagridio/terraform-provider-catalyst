output "oneday_api_key_token" {
  value       = catalyst_service_account_api_key.oneday_api_key.token
  sensitive   = true
  description = "API key token"
}

output "oneday_api_key_expires_at" {
  value       = catalyst_service_account_api_key.oneday_api_key.expires_at
  description = "API key expiration"
}

output "permanent_api_key_token" {
  value       = catalyst_service_account_api_key.permanent_api_key.token
  sensitive   = true
  description = "API key token"
}

output "permanent_api_key_expires_at" {
  value       = catalyst_service_account_api_key.permanent_api_key.expires_at
  description = "API key expiration"
}

