variable "aws_region" {
  type        = string
  description = "The AWS region to deploy the infrastructure"
  default     = "ap-south-1"
}

variable "environment" {
  type        = string
  description = "Deployment environment (e.g. dev, staging, production)"
  default     = "production"
}