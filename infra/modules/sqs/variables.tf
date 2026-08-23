variable "name" {
  type        = string
  description = "Name of the SQS queue (DLQ is named <name>-dlq unless dlq_name is set)"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "shared"
}

variable "allowed_task_role_name" {
  type        = string
  description = "IAM role name allowed to use this queue (execution-api task role). Queue policies allow only this principal."
  nullable    = false
}

variable "dlq_name" {
  type        = string
  description = "Name of the dead-letter queue"
  default     = null
  nullable    = true
}

variable "max_receive_count" {
  type        = number
  description = "Receives before a message is moved to the DLQ"
  default     = 5
}

variable "message_retention_seconds" {
  type        = number
  description = "How long SQS retains messages on the source queue (max 1209600 = 14 days)"
  default     = 1209600
}

variable "dlq_message_retention_seconds" {
  type        = number
  description = "How long SQS retains parked messages on the DLQ (max 1209600 = 14 days)"
  default     = 1209600
}

variable "visibility_timeout_seconds" {
  type        = number
  description = "Default visibility timeout. Workers extend per message with ChangeMessageVisibility."
  default     = 60
}

variable "receive_wait_time_seconds" {
  type        = number
  description = "Long-poll wait for ReceiveMessage (0–20 seconds)"
  default     = 20
}
