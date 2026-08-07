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
