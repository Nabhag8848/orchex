variable "name" {
  type        = string
  description = "Lambda function name"
  nullable    = false
}

variable "invoker_task_role_name" {
  type        = string
  description = "IAM role name allowed to Invoke — execution-worker task role (pinned name, not a module output, to avoid a cycle)"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "execution-api"
}
