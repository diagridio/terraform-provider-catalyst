terraform {
  required_providers {
    catalyst = {
      source  = "diagridio/catalyst"
      version = "~> 0.1"
    }
  }
}

provider "catalyst" {
  # Configuration options
  api_key  = var.api_key
  endpoint = var.endpoint
}