# Catalyst Provider Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or ar testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** example file for the provider index page
* **data-sources/`full data source name`/data-source.tf** example file for the named data source page
* **resources/`full resource name`/resource.tf** example file for the named resource page

## Available Examples

### Service Account Examples
- **resources/catalyst_service_account/** - Service account resource examples
- **data-sources/catalyst_service_account/** - Service account data source examples

### Comprehensive Examples
- **comprehensive/** - Complete integration examples showing service accounts with other Catalyst resources

## Quick Start

Each example directory contains:
- Terraform configuration files (*.tf)
- README.md with detailed usage instructions
- terraform.tfvars.example with sample variable values
- outputs.tf showing useful output patterns

To use any example:
1. Navigate to the example directory
2. Copy terraform.tfvars.example to terraform.tfvars
3. Edit terraform.tfvars with your actual values
4. Run `terraform init && terraform apply`

## Service Account Features

The service account examples demonstrate:
- ✅ Creating service accounts with different roles (admin/viewer)
- ✅ Looking up existing service accounts by name or UID
- ✅ Integration patterns with other Catalyst resources
- ✅ Output patterns for external tool integration
- ✅ Security best practices and role-based access
