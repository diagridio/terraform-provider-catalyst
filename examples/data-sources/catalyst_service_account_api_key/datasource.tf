# Look up service account by name
data "catalyst_service_account" "basic_service_account" {
  name = "basic-service-account"
}

# Look up service account API key by service account ID and token ID
data "catalyst_service_account_api_key" "api_key" {
  name               = "basic-service-account"
  service_account_id = data.catalyst_service_account.basic_service_account.email
  token_id           = "token-id"
}

