# Service Account API Key Data Source Examples

This directory contains examples for using the `catalyst_service_account_api_key` data source.

## Example Usage

### Basic Data Source Lookup

```hcl
data "catalyst_service_account_api_key" "existing_api_key" {
  service_account_id = "your-service-account-id"
  token_id          = "your-token-id"
}
```

### Using with Resources

```hcl
# Create a service account and API key
resource "catalyst_service_account" "example" {
  name        = "example-service-account"
  description = "Example service account"
  role        = "viewer"
}

resource "catalyst_service_account_api_key" "example_key" {
  name               = "example-api-key"
  service_account_id = catalyst_service_account.example.email
  expire_in_seconds  = 86400 # 24 hours
}

# Look up the created API key
data "catalyst_service_account_api_key" "example_lookup" {
  service_account_id = catalyst_service_account_api_key.example_key.service_account_id
  token_id          = catalyst_service_account_api_key.example_key.token_id
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

## Data Source Attributes

The data source provides the following attributes:

### Required Arguments

- `service_account_id` - (Required) The ID of the service account that owns the API key
- `token_id` - (Required) The token ID of the API key to look up

### Exported Attributes

- `name` - The name of the API key
- `expire_in_seconds` - The expiration time in seconds (if set)
- `token` - The actual API key token (sensitive)
- `created_at` - The creation timestamp

## Important Notes

### Security Considerations

- **Sensitive Data**: The `token` attribute contains the actual API key and is marked as sensitive
- **State File**: The API key token will be stored in your Terraform state file. Ensure your state file is properly secured

### Usage Patterns

- **Validation**: Use data sources to validate that API keys exist before using them in other resources
- **Reference**: Reference existing API keys in other configurations without managing them directly
- **Lookup**: Look up API key metadata for reporting or conditional logic

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

- `api_key_info`: Non-sensitive API key metadata
- `api_key_token`: The sensitive API key token (use with caution)
- `example_lookup_info`: Information from the example lookup

## Use Cases

### Conditional Resource Creation

```hcl
data "catalyst_service_account_api_key" "check" {
  service_account_id = var.service_account_id
  token_id          = var.token_id
}

# Use the data source in conditional logic
locals {
  api_key_exists = data.catalyst_service_account_api_key.check.token_id != null
}
```

### Integration with Other Resources

```hcl
# Use API key information in other resources
resource "local_file" "api_key_info" {
  content = jsonencode({
    name       = data.catalyst_service_account_api_key.existing_api_key.name
    created_at = data.catalyst_service_account_api_key.existing_api_key.created_at
  })
  filename = "api_key_info.json"
}
```