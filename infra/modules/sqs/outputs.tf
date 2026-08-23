output "queue_arn" {
  description = "ARN of the source queue"
  value       = module.this.queue_arn
}

output "queue_url" {
  description = "URL of the source queue"
  value       = module.this.queue_url
}

output "queue_name" {
  description = "Name of the source queue"
  value       = module.this.queue_name
}

output "dead_letter_queue_arn" {
  description = "ARN of the dead-letter queue"
  value       = module.this.dead_letter_queue_arn
}

output "dead_letter_queue_url" {
  description = "URL of the dead-letter queue"
  value       = module.this.dead_letter_queue_url
}

output "dead_letter_queue_name" {
  description = "Name of the dead-letter queue"
  value       = module.this.dead_letter_queue_name
}

output "allowed_task_role_arn" {
  description = "IAM role ARN allowed by the queue resource policies"
  value       = local.allowed_task_role_arn
}

output "queue_actions" {
  description = "SQS actions granted to the execution-api task role"
  value       = local.queue_actions
}
