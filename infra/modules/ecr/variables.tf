variable "repository_name" {
  type        = string
  description = "The name of the ECR repository"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  nullable    = false
}
