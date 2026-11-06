# Service Account API Key Resource Examples

This directory contains examples for using the `catalyst_service_account_api_key` resource.

## Example Usage

### Basic API Key

```hcl
resource "catalyst_service_account_api_key" "basic_api_key" {
  name               = "basic-api-key"
  service_account_id = "your-service-account-id"
  expire_in_seconds  = 86400 # 24 hours
}
```

### API Key with Different Expiration Times

```hcl
# Long-lived API key (30 days)
resource "catalyst_service_account_api_key" "long_lived_api_key" {
  name               = "long-lived-api-key"
  service_account_id = "your-service-account-id"
  expire_in_seconds  = 2592000 # 30 days
}

# Permanent API key (no expiration)
resource "catalyst_service_account_api_key" "permanent_api_key" {
  name               = "permanent-api-key"
  service_account_id = "your-service-account-id"
  # expire_in_seconds is optional - if not set, the key won't expire
}
```

### Using with Service Account Resource

```hcl
resource "catalyst_service_account" "example" {
  name        = "example-service-account"
  description = "Example service account for API key"
  role        = "admin"
}

resource "catalyst_service_account_api_key" "example_api_key" {
  name               = "example-api-key"
  service_account_id = catalyst_service_account.example.email
  expire_in_seconds  = 604800 # 7 days
}
```

## Required Variables

Before running these examples, make sure to set the required variables:

1. Create a `terraform.tfvars` file:
   ```hcl
   api_key = "your-catalyst-api-key"
   endpoint = "https://api.r1.diagrid.io"  # optional, defaults to this value
   ```

2. Or set environment variables:
   ```bash
   export CATALYST_API_KEY="your-catalyst-api-key"
   export CATALYST_API_ENDPOINT="https://api.r1.diagrid.io"  # optional
   ```

## Important Notes

### Security Considerations

- **API Key Token**: The `token` attribute contains the actual API key and is marked as sensitive. Be careful when outputting this value.
- **State File**: The API key token will be stored in your Terraform state file. Ensure your state file is properly secured.
- **Expiration**: Set appropriate expiration times for your API keys based on your security requirements.

### Resource Behavior

- **Immutable**: API keys cannot be updated. Any change to the name, service_account_id, or expire_in_seconds will force a new resource to be created.
- **Import**: You can import existing API keys using their token_id.
- **Deletion**: When the resource is destroyed, the API key will be permanently deleted from the service account.

## Running the Examples

1. Clone this repository and navigate to this directory
2. Set up your variables (see above)
3. Initialize Terraform:
   ```bash
   terraform init
   ```
4. Plan your changes:
   ```bash
   terraform plan
   ```
5. Apply the configuration:
   ```bash
   terraform apply
   ```

## Outputs

The examples include several outputs to demonstrate different ways to access API key information:

- `basic_api_key`: Non-sensitive API key metadata
- `basic_api_key_token`: The sensitive API key token (use with caution)
- `api_key_summary`: Summary of all created API keys

## Import

To import an existing API key:

```bash
terraform import catalyst_service_account_api_key.example <token_id>
```

Replace `<token_id>` with the actual token ID of the API key you want to import.