variable "name" {
  type        = string
  description = "Secrets Manager secret name (e.g. orchex/DATABASE_URL)"
  nullable    = false
}

variable "secret_string" {
  type        = string
  description = "Secret payload stored as the current secret version"
  sensitive   = true
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "shared"
}

variable "description" {
  type        = string
  description = "Human-readable description of the secret"
  default     = null
  nullable    = true
}
