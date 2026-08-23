variable "name" {
  type        = string
  description = "Name of the ECS service"
  nullable    = false
}

variable "cluster_arn" {
  type        = string
  description = "ARN of the ECS cluster to run this service on"
  nullable    = false
}

variable "container_name" {
  type        = string
  description = "Name of the container inside the task"
  nullable    = false
}

variable "image" {
  type        = string
  description = "Container image URI (ECR repository URL with tag)"
  nullable    = false
}

variable "container_port" {
  type        = number
  description = "Port the container listens on"
  default     = 8080
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

variable "desired_count" {
  type        = number
  description = "Number of tasks to keep running"
  default     = 1
}

variable "target_group_arn" {
  type        = string
  description = "ALB target group ARN; ECS registers task IPs automatically"
  nullable    = false
}

variable "alb_security_group_id" {
  type        = string
  description = "ALB security group allowed to reach the container port"
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
  nullable    = false
}

variable "database_url_secret_arn" {
  type        = string
  description = "Secrets Manager ARN for DATABASE_URL; injected into the container when set"
  default     = null
  nullable    = true
}

variable "extra_environment" {
  type = list(object({
    name  = string
    value = string
  }))
  description = "Additional container environment variables (merged after HTTP_ADDR)"
  default     = []
  nullable    = false
}

variable "tasks_iam_role_name" {
  type        = string
  description = "Name of the ECS task role. Set with tasks_iam_role_use_name_prefix = false for a stable ARN (queue policies)."
  default     = null
  nullable    = true
}

variable "tasks_iam_role_use_name_prefix" {
  type        = bool
  description = "When true, AWS appends a unique suffix to the task role name"
  default     = true
  nullable    = false
}

variable "tasks_iam_role_statements" {
  type = list(object({
    sid           = optional(string)
    actions       = optional(list(string))
    not_actions   = optional(list(string))
    effect        = optional(string)
    resources     = optional(list(string))
    not_resources = optional(list(string))
    principals = optional(list(object({
      type        = string
      identifiers = list(string)
    })))
    not_principals = optional(list(object({
      type        = string
      identifiers = list(string)
    })))
    condition = optional(list(object({
      test     = string
      values   = list(string)
      variable = string
    })))
  }))
  description = "Extra IAM statements attached to the ECS task role"
  default     = null
}
