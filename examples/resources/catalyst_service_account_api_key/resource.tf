# Basic service account resource example
resource "catalyst_service_account" "basic_service_account" {
  name        = "basic-service-account"
  description = "Service account for automated tasks"
  owner       = "engineering@example.com"
  role        = "cra.diagrid:admin"
}

# Basic service account API key
resource "catalyst_service_account_api_key" "oneday_api_key" {
  name               = "basic-api-key"
  service_account_id = catalyst_service_account.basic_service_account.name
  expire_in_seconds  = 86400 # 24 hours
}

# API key without expiration (permanent)
resource "catalyst_service_account_api_key" "permanent_api_key" {
  name               = "permanent-api-key"
  service_account_id = catalyst_service_account.basic_service_account.name
  # expire_in_seconds is optional - if not set, the key won't expire
}

