# Specify required provider as maintained
terraform {
  required_providers {
    catalyst = {
      source = "diagridio/catalyst"
    }
  }
}

provider "catalyst" {
  api_key  = var.api_key
  endpoint = var.endpoint
}

