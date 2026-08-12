output "arn" {
  description = "Secret ARN for ECS task definition secrets and task_exec_secret_arns"
  value       = module.this.secret_arn
}

output "id" {
  description = "Secret ID"
  value       = module.this.secret_id
}

output "name" {
  description = "Secret name"
  value       = module.this.secret_name
}
