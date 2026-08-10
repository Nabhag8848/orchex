variable "name" {
  type        = string
  description = "Name of the ALB"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "shared"
}
