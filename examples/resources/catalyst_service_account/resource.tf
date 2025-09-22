# Basic service account resource example
resource "catalyst_service_account" "basic_service_account" {
  name        = "basic-service-account"
  description = "Service account for automated tasks"
  owner       = "engineering@example.com"
  role        = "cra.diagrid:admin"
}

# Service account with different role
resource "catalyst_service_account" "viewer_service_account" {
  name        = "readonly-service-account"
  description = "Read-only service account for monitoring"
  owner       = "ops@example.com"
  role        = "cra.diagrid:viewer"
}

# Service account for CI/CD pipeline
resource "catalyst_service_account" "cicd_service_account" {
  name        = "cicd-automation"
  description = "Service account for CI/CD deployment automation"
  owner       = "devops@example.com"
  role        = "cra.diagrid:admin"
}
