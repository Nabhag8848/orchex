output "db_instance_endpoint" {
  description = "The connection endpoint"
  value       = module.this.db_instance_endpoint
}

output "db_instance_address" {
  description = "The address of the RDS instance"
  value       = module.this.db_instance_address
}

output "db_instance_port" {
  description = "The database port"
  value       = module.this.db_instance_port
}

output "db_instance_name" {
  description = "The database name"
  value       = module.this.db_instance_name
}

output "db_instance_identifier" {
  description = "The RDS instance identifier"
  value       = module.this.db_instance_identifier
}

output "db_instance_arn" {
  description = "The ARN of the RDS instance"
  value       = module.this.db_instance_arn
}

output "db_instance_status" {
  description = "The RDS instance status"
  value       = module.this.db_instance_status
}

output "db_instance_engine" {
  description = "The database engine"
  value       = module.this.db_instance_engine
}

output "db_instance_engine_version_actual" {
  description = "The running version of the database"
  value       = module.this.db_instance_engine_version_actual
}

output "db_subnet_group_id" {
  description = "The db subnet group name"
  value       = aws_db_subnet_group.this.id
}

output "db_subnet_group_arn" {
  description = "The ARN of the db subnet group"
  value       = aws_db_subnet_group.this.arn
}

output "security_group_id" {
  description = "RDS security group ID"
  value       = aws_security_group.rds.id
}

output "security_group_arn" {
  description = "RDS security group ARN"
  value       = aws_security_group.rds.arn
}

output "db_instance_master_user_secret_arn" {
  description = "ARN of the RDS-managed master user secret"
  value       = module.this.db_instance_master_user_secret_arn
  sensitive   = true
}
