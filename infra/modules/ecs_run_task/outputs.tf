output "task_definition_arn" {
  description = "Full ARN of the task definition (family and revision)"
  value       = module.this.task_definition_arn
}

output "task_definition_family" {
  description = "Task definition family name for aws ecs run-task"
  value       = module.this.task_definition_family
}

output "task_definition_revision" {
  description = "Task definition revision"
  value       = module.this.task_definition_revision
}

output "task_exec_iam_role_arn" {
  description = "Task execution IAM role ARN"
  value       = module.this.task_exec_iam_role_arn
}

output "security_group_id" {
  description = "Security group attached to the task ENI (egress)"
  value       = module.this.security_group_id
}

output "subnet_ids" {
  description = "Default VPC subnet IDs for run-task network configuration"
  value       = data.aws_subnets.default.ids
}

output "run_task_network_configuration" {
  description = "awsvpcConfiguration for aws ecs run-task (comma-separated lists)"
  value = {
    subnets          = join(",", data.aws_subnets.default.ids)
    security_groups  = join(",", [module.this.security_group_id, var.data_plane_client_security_group_id])
    assign_public_ip = "ENABLED"
  }
}
