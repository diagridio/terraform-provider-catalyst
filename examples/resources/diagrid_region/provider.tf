# Specify required provider as maintained
terraform {
  required_providers {
    diagrid = {
      source  = "diagridio/diagrid"
      version = "~> 0.0.1"
    }
  }
}

provider "diagrid" {
  api_key  = var.api_key
  endpoint = var.endpoint
}

