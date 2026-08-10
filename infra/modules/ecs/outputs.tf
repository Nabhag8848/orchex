output "arn" {
  description = "ARN that identifies the cluster"
  value       = module.this.arn
}

output "id" {
  description = "ID that identifies the cluster"
  value       = module.this.id
}

output "name" {
  description = "Name that identifies the cluster"
  value       = module.this.name
}

output "cloudwatch_log_group_name" {
  description = "Name of CloudWatch log group created"
  value       = module.this.cloudwatch_log_group_name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of CloudWatch log group created"
  value       = module.this.cloudwatch_log_group_arn
}

output "cluster_capacity_providers" {
  description = "Map of cluster capacity providers attributes"
  value       = module.this.cluster_capacity_providers
}

output "capacity_providers" {
  description = "Map of autoscaling capacity providers created and their attributes"
  value       = module.this.capacity_providers
}

output "task_exec_iam_role_arn" {
  description = "Task execution IAM role ARN"
  value       = module.this.task_exec_iam_role_arn
}

output "task_exec_iam_role_name" {
  description = "Task execution IAM role name"
  value       = module.this.task_exec_iam_role_name
}

output "task_exec_iam_role_unique_id" {
  description = "Stable and unique string identifying the task execution IAM role"
  value       = module.this.task_exec_iam_role_unique_id
}

output "infrastructure_iam_role_arn" {
  description = "The Amazon Resource Name (ARN) specifying the IAM role"
  value       = module.this.infrastructure_iam_role_arn
}

output "infrastructure_iam_role_name" {
  description = "IAM role name"
  value       = module.this.infrastructure_iam_role_name
}

output "infrastructure_iam_role_unique_id" {
  description = "Stable and unique string identifying the IAM role"
  value       = module.this.infrastructure_iam_role_unique_id
}

output "node_iam_role_arn" {
  description = "The Amazon Resource Name (ARN) specifying the IAM role"
  value       = module.this.node_iam_role_arn
}

output "node_iam_role_name" {
  description = "IAM role name"
  value       = module.this.node_iam_role_name
}

output "node_iam_role_unique_id" {
  description = "Stable and unique string identifying the IAM role"
  value       = module.this.node_iam_role_unique_id
}

output "node_iam_instance_profile_arn" {
  description = "ARN assigned by AWS to the instance profile"
  value       = module.this.node_iam_instance_profile_arn
}

output "node_iam_instance_profile_id" {
  description = "Instance profile's ID"
  value       = module.this.node_iam_instance_profile_id
}

output "node_iam_instance_profile_unique" {
  description = "Stable and unique string identifying the IAM instance profile"
  value       = module.this.node_iam_instance_profile_unique
}

output "data_plane_client_security_group_id" {
  description = "Shared client security group attached to ECS task ENIs for data plane access"
  value       = aws_security_group.data_plane_client.id
}

output "data_plane_client_security_group_arn" {
  description = "ARN of the shared ECS data plane client security group"
  value       = aws_security_group.data_plane_client.arn
}
