# Set the variable value in *.tfvars file or using -var="api_key=..." CLI flag
variable "api_key" {
  type        = string
  sensitive   = true
  description = "Diagrid API key"
}

variable "endpoint" {
  type        = string
  description = "Diagrid API endpoint"
}

variable "organization_id" {
  type        = string
  description = "Diagrid organization ID"
}

variable "project_name" {
  type        = string
  description = "Diagrid project name"
}
