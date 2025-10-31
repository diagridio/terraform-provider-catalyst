# Catalyst Provider Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or are testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** - Example file for the provider index page
* **data-sources/`full data source name`/datasource.tf** - Example file for the named data source page
* **resources/`full resource name`/resource.tf** - Example file for the named resource page

## Quick Start

Each example directory contains:
- **provider.tf** - Provider configuration
- **resource.tf** or **datasource.tf** - Main resource/data source configuration
- **outputs.tf** - Output values demonstrating available attributes

To use any example:
1. Navigate to the example directory
2. Update the provider configuration in `provider.tf` with your credentials
3. Modify the resource/data source configuration as needed
4. Run `terraform init && terraform apply`

