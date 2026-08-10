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
