# Catalyst Service Account Resource Examples

This directory contains examples for using the `catalyst_service_account` resource.

## Prerequisites

- Terraform >= 0.13
- Valid Catalyst API key
- Access to Catalyst API endpoint

## Files

- `resource.tf` - Service account resource definitions
- `provider.tf` - Provider configuration
- `variables.tf` - Variable definitions
- `outputs.tf` - Output definitions
- `terraform.tfvars.example` - Example variable values

## Usage

1. Copy the example variables file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your actual values:
   ```hcl
   api_key = "your-actual-api-key"
   endpoint = "https://api.diagrid.io"  # or your custom endpoint
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Examples Included

### 1. Basic Service Account
Creates a basic service account with admin role:
```hcl
resource "catalyst_service_account" "basic_service_account" {
  name        = "my-service-account"
  description = "Service account for automated tasks"
  owner       = "engineering@example.com"
  role        = "admin"
}
```

### 2. Viewer Service Account
Creates a read-only service account:
```hcl
resource "catalyst_service_account" "viewer_service_account" {
  name        = "readonly-service-account"
  description = "Read-only service account for monitoring"
  owner       = "ops@example.com"
  role        = "viewer"
}
```

### 3. CI/CD Service Account
Creates a service account for automation:
```hcl
resource "catalyst_service_account" "cicd_service_account" {
  name        = "cicd-automation"
  description = "Service account for CI/CD deployment automation"
  owner       = "devops@example.com"
  role        = "admin"
}
```

## Available Attributes

### Required
- `name` - Name of the service account (must be unique)
- `description` - Description of the service account
- `owner` - Email of the service account owner
- `role` - Role for the service account (`admin` or `viewer`)

### Computed
- `uid` - Unique identifier of the service account
- `email` - Email address of the service account
- `status` - Current status of the service account
- `updated_at` - Last update timestamp

## Import

Service accounts can be imported using their UID:
```bash
terraform import catalyst_service_account.example "service-account-uid"
```

## Notes

- Service account names must be unique within your organization
- Changing the `name` attribute will force resource replacement
- The `email`, `status`, and `updated_at` attributes are computed by the API
- Service accounts are created with an `active` status by default