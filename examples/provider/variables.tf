# Set the variable value in *.tfvars file or using -var="api_key=..." CLI flag
variable "api_key" {
  type        = string
  sensitive   = true
  description = "Catalyst API key"
}

variable "endpoint" {
  type        = string
  description = "Catalyst API endpoint"
}

variable "organization_id" {
  type        = string
  description = "Catalyst organization ID"
}

variable "project_name" {
  type        = string
  description = "Catalyst project name"
}
