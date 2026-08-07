output "id" {
  description = "ARN that identifies the service"
  value       = module.this.id
}

output "name" {
  description = "Name of the service"
  value       = module.this.name
}

output "iam_role_name" {
  description = "Service IAM role name"
  value       = module.this.iam_role_name
}

output "iam_role_arn" {
  description = "Service IAM role ARN"
  value       = module.this.iam_role_arn
}

output "iam_role_unique_id" {
  description = "Stable and unique string identifying the service IAM role"
  value       = module.this.iam_role_unique_id
}

output "container_definitions" {
  description = "Container definitions"
  value       = module.this.container_definitions
}

output "task_definition_arn" {
  description = "Full ARN of the Task Definition (including both family and revision)"
  value       = module.this.task_definition_arn
}

output "task_definition_revision" {
  description = "Revision of the task in a particular family"
  value       = module.this.task_definition_revision
}

output "task_definition_family" {
  description = "The unique name of the task definition"
  value       = module.this.task_definition_family
}

output "task_exec_iam_role_name" {
  description = "Task execution IAM role name"
  value       = module.this.task_exec_iam_role_name
}

output "task_exec_iam_role_arn" {
  description = "Task execution IAM role ARN"
  value       = module.this.task_exec_iam_role_arn
}

output "task_exec_iam_role_unique_id" {
  description = "Stable and unique string identifying the task execution IAM role"
  value       = module.this.task_exec_iam_role_unique_id
}

output "tasks_iam_role_name" {
  description = "Tasks IAM role name"
  value       = module.this.tasks_iam_role_name
}

output "tasks_iam_role_arn" {
  description = "Tasks IAM role ARN"
  value       = module.this.tasks_iam_role_arn
}

output "tasks_iam_role_unique_id" {
  description = "Stable and unique string identifying the tasks IAM role"
  value       = module.this.tasks_iam_role_unique_id
}

output "task_set_id" {
  description = "The ID of the task set"
  value       = module.this.task_set_id
}

output "task_set_arn" {
  description = "The Amazon Resource Name (ARN) that identifies the task set"
  value       = module.this.task_set_arn
}

output "task_set_stability_status" {
  description = "The stability status. This indicates whether the task set has reached a steady state"
  value       = module.this.task_set_stability_status
}

output "task_set_status" {
  description = "The status of the task set"
  value       = module.this.task_set_status
}

output "autoscaling_policies" {
  description = "Map of autoscaling policies and their attributes"
  value       = module.this.autoscaling_policies
}

output "autoscaling_scheduled_actions" {
  description = "Map of autoscaling scheduled actions and their attributes"
  value       = module.this.autoscaling_scheduled_actions
}

output "security_group_arn" {
  description = "Amazon Resource Name (ARN) of the security group"
  value       = module.this.security_group_arn
}

output "security_group_id" {
  description = "ID of the security group"
  value       = module.this.security_group_id
}

output "infrastructure_iam_role_arn" {
  description = "Infrastructure IAM role ARN"
  value       = module.this.infrastructure_iam_role_arn
}

output "infrastructure_iam_role_name" {
  description = "Infrastructure IAM role name"
  value       = module.this.infrastructure_iam_role_name
}
