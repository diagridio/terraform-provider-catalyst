# Specify required provider as maintained
terraform {
  required_providers {
    diagrid = {
      source = "diagridio/diagrid"
    }
  }
}

provider "diagrid" {
  api_key  = var.api_key
  endpoint = var.endpoint
}

