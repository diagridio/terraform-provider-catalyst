# Terraform Provider for Catalyst

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

The Catalyst provider enables you to manage Diagrid Catalyst resources using Terraform. You'll need an API key to authenticate with the Catalyst API.

### Provider Configuration

```hcl
terraform {
  required_providers {
    catalyst = {
      source = "diagridio/catalyst"
    }
  }
}

provider "catalyst" {
  api_key  = var.api_key
  endpoint = var.endpoint  # Optional, defaults to production endpoint
}
```

### Authentication

Set your API key as a variable or environment variable:

```bash
export TF_VAR_api_key="your-catalyst-api-key"
```

### Available Resources

#### Regions
Create and manage Catalyst regions:

```hcl
resource "catalyst_region" "us_west" {
  name     = "us-west-region"
  ingress  = "https://*.example.com:443"
  host     = "us-west-host"
  location = "us-west-1"
}
```

#### Projects
Create projects within regions:

```hcl
resource "catalyst_project" "my_project" {
  region = catalyst_region.us_west.name
  name   = "my-application"
}
```

#### Service Accounts
Manage service accounts for programmatic access:

```hcl
resource "catalyst_service_account" "automation" {
  name        = "automation-account"
  description = "Service account for CI/CD automation"
  owner       = "devops@example.com"
  role        = "cra.diagrid:admin"  # or "cra.diagrid:viewer"
}
```

#### Service Account API Keys
Generate API keys for service accounts:

```hcl
resource "catalyst_service_account_api_key" "automation_key" {
  name               = "automation-api-key"
  service_account_id = catalyst_service_account.automation.name
  expire_in_seconds  = 86400  # 24 hours (optional)
}

# Access the generated token
output "api_key_token" {
  value     = catalyst_service_account_api_key.automation_key.token
  sensitive = true
}
```

### Available Data Sources

All resources are also available as data sources for referencing existing infrastructure:

- `data.catalyst_organization` - Organization information
- `data.catalyst_region` - Region details
- `data.catalyst_project` - Project information
- `data.catalyst_service_account` - Service account details
- `data.catalyst_service_account_api_key` - API key information

### Example Usage

See the `examples/` directory for complete working examples of each resource type.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
