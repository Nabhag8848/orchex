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

output "producer_task_role_arn" {
  description = "IAM role ARN allowed to produce (execution-api)"
  value       = local.producer_task_role_arn
}

output "consumer_task_role_arn" {
  description = "IAM role ARN allowed to consume (worker); null when unset"
  value       = local.consumer_task_role_arn
}

output "producer_queue_actions" {
  description = "SQS actions for the execution-api task role"
  value       = local.producer_queue_actions
}

output "consumer_queue_actions" {
  description = "SQS actions for the worker task role (source queue only)"
  value       = local.consumer_queue_actions
}
