# Specify required provider as maintained
terraform {
  required_providers {
    catalyst = {
      source  = "diagridio/catalyst"
      version = "~> 0.0.1"
    }
  }
}

provider "catalyst" {
  api_key  = var.api_key
  endpoint = var.endpoint
}

