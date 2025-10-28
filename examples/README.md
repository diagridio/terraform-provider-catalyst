# Catalyst Provider Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or are testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** - Example file for the provider index page
* **data-sources/`full data source name`/datasource.tf** - Example file for the named data source page
* **resources/`full resource name`/resource.tf** - Example file for the named resource page

## Available Examples

### Resources

- **resources/catalyst_project/** - Project resource examples
- **resources/catalyst_region/** - Region resource examples
- **resources/catalyst_service_account/** - Service account resource examples
- **resources/catalyst_appid/** - App identity resource examples
- **resources/catalyst_component/** - Dapr component resource examples
- **resources/catalyst_pubsub/** - PubSub resource examples
- **resources/catalyst_kvstore/** - Key-value store resource examples
- **resources/catalyst_subscription/** - Dapr subscription resource examples
- **resources/catalyst_resiliency/** - Dapr resiliency policy resource examples
- **resources/catalyst_configuration/** - Dapr configuration resource examples

### Data Sources

- **data-sources/catalyst_organization/** - Organization data source examples
- **data-sources/catalyst_project/** - Project data source examples
- **data-sources/catalyst_region/** - Region data source examples
- **data-sources/catalyst_service_account/** - Service account data source examples
- **data-sources/catalyst_appid/** - App identity data source examples
- **data-sources/catalyst_component/** - Dapr component data source examples
- **data-sources/catalyst_pubsub/** - PubSub data source examples
- **data-sources/catalyst_kvstore/** - Key-value store data source examples
- **data-sources/catalyst_subscription/** - Dapr subscription data source examples
- **data-sources/catalyst_resiliency/** - Dapr resiliency policy data source examples
- **data-sources/catalyst_configuration/** - Dapr configuration data source examples

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

## Resource Categories

### Infrastructure Resources
- **Project** - Manage Catalyst projects
- **Region** - Manage Catalyst regions
- **Service Account** - Manage service accounts with different roles
- **App Identity** - Manage application identities
- **PubSub** - Manage pub/sub messaging infrastructure
- **KVStore** - Manage key-value stores

### Dapr Resources
- **Component** - Manage Dapr components (state stores, pub/sub, etc.)
- **Subscription** - Manage Dapr subscriptions to topics
- **Resiliency** - Manage Dapr resiliency policies
- **Configuration** - Manage Dapr configurations
