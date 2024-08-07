# Set the variable value in *.tfvars file or using -var="api_key=..." CLI flag
variable "api_key" {
  type        = string
  sensitive   = true
  description = "Catalyst API key"
}

variable "endpoint" {
  type        = string
  description = "Catalyst API endpoint"
  default     = "https://api.diagrid.io"
}

