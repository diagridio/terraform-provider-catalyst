## 0.1.0 (Unreleased)

FEATURES:

**New Resources:**
* `catalyst_region` - Manage Catalyst regions with ingress, host, and location configuration
* `catalyst_project` - Create and manage projects within regions  
* `catalyst_service_account` - Manage service accounts with role-based access control
* `catalyst_service_account_api_key` - Generate and manage API keys for service accounts with optional expiration

**New Data Sources:**
* `catalyst_organization` - Retrieve organization information
* `catalyst_region` - Fetch details about existing regions
* `catalyst_project` - Query project information
* `catalyst_service_account` - Access service account details
* `catalyst_service_account_api_key` - Retrieve API key information

**Provider Configuration:**
* API key authentication support
* Configurable endpoint for different environments
* Full Terraform state management for all resources

**Documentation:**
* Comprehensive README with usage examples
* Generated documentation for all resources and data sources
* Complete example configurations in `/examples` directory
