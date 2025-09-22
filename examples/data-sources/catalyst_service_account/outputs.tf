# Output specific attributes for common use cases
output "service_account" {
  value = {
    name        = data.catalyst_service_account.basic_service_account.name
    email       = data.catalyst_service_account.basic_service_account.email
    description = data.catalyst_service_account.basic_service_account.description
    owner       = data.catalyst_service_account.basic_service_account.owner
    role        = data.catalyst_service_account.basic_service_account.role
  }
  description = "Detailed service account information"
}

