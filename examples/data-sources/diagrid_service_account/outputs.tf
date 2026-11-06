# Output specific attributes for common use cases
output "service_account" {
  value = {
    name        = data.diagrid_service_account.basic_service_account.name
    email       = data.diagrid_service_account.basic_service_account.email
    description = data.diagrid_service_account.basic_service_account.description
    owner       = data.diagrid_service_account.basic_service_account.owner
    role        = data.diagrid_service_account.basic_service_account.role
  }
  description = "Detailed service account information"
}

