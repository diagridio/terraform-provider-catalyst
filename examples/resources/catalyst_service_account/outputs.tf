output "basic_service_account" {
  value       = catalyst_service_account.basic_service_account
  description = "Basic service account details"
}

output "viewer_service_account" {
  value       = catalyst_service_account.viewer_service_account
  description = "Viewer service account details"
}

output "cicd_service_account" {
  value       = catalyst_service_account.cicd_service_account
  description = "CI/CD service account details"
}

output "basic_service_account_email" {
  value       = catalyst_service_account.basic_service_account.email
  description = "Email of the basic service account"
}

output "service_account_summary" {
  value = {
    basic_sa = {
      name  = catalyst_service_account.basic_service_account.name
      email = catalyst_service_account.basic_service_account.email
      role  = catalyst_service_account.basic_service_account.role
    }
    viewer_sa = {
      name  = catalyst_service_account.viewer_service_account.name
      email = catalyst_service_account.viewer_service_account.email
      role  = catalyst_service_account.viewer_service_account.role
    }
    cicd_sa = {
      name  = catalyst_service_account.cicd_service_account.name
      email = catalyst_service_account.cicd_service_account.email
      role  = catalyst_service_account.cicd_service_account.role
    }
  }
  description = "Summary of all service accounts"
}
