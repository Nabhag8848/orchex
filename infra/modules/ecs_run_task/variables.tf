variable "name" {
  type        = string
  description = "Name prefix for the ECS task definition family"
  nullable    = false
}

variable "cluster_arn" {
  type        = string
  description = "ARN of the ECS cluster tasks run on"
  nullable    = false
}

variable "container_name" {
  type        = string
  description = "Name of the container inside the task"
  default     = "migrate"
}

variable "image" {
  type        = string
  description = "Container image URI (ECR repository URL with tag)"
  nullable    = false
}

variable "cpu" {
  type        = number
  description = "Fargate task CPU units"
  default     = 256
}

variable "memory" {
  type        = number
  description = "Fargate task memory (MiB)"
  default     = 512
}

variable "database_url_secret_arn" {
  type        = string
  description = "Secrets Manager ARN for the shared DATABASE_URL (injected as GOOSE_DBSTRING)"
  nullable    = false
}

variable "data_plane_client_security_group_id" {
  type        = string
  description = "Shared ECS data plane client security group attached to task ENIs for RDS access"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "shared"
}
