variable "name" {
  type        = string
  description = "Name prefix for RDS resources"
  nullable    = false
}

variable "service" {
  type        = string
  description = "Logical service name for tagging"
  default     = "shared"
}

variable "db_name" {
  type        = string
  description = "Name of the default database to create"
  default     = "orchex"
}

variable "username" {
  type        = string
  description = "Master username for the database"
  default     = "orchex"
}

variable "engine" {
  type        = string
  description = "Database engine"
  default     = "postgres"
}

variable "engine_version" {
  type        = string
  description = "Database engine version"
  default     = "17"
}

variable "family" {
  type        = string
  description = "DB parameter group family"
  default     = "postgres17"
}

variable "major_engine_version" {
  type        = string
  description = "Major engine version used by option/parameter groups"
  default     = "17"
}

variable "instance_class" {
  type        = string
  description = "RDS instance class (db.t4g.medium = 2 vCPU, 4 GiB)"
  default     = "db.t4g.medium"
}

variable "allocated_storage" {
  type        = number
  description = "Allocated storage in GiB (20 is the RDS minimum for gp3 Postgres)"
  default     = 20
}

variable "port" {
  type        = number
  description = "Database port"
  default     = 5432
}

variable "publicly_accessible" {
  type        = bool
  description = "Whether the instance gets a public IP"
  default     = false
}

variable "multi_az" {
  type        = bool
  description = "Deploy a standby instance in another AZ"
  default     = false
}

variable "backup_retention_period" {
  type        = number
  description = "Days to retain automated backups (0 disables backups)"
  default     = 0
}

variable "storage_encrypted" {
  type        = bool
  description = "Whether to encrypt storage at rest"
  default     = true
}

variable "deletion_protection" {
  type        = bool
  description = "Prevent accidental deletion via the AWS console or API"
  default     = false
}

variable "skip_final_snapshot" {
  type        = bool
  description = "Skip a final snapshot when the instance is destroyed"
  default     = true
}

variable "data_plane_client_security_group_id" {
  type        = string
  description = "ECS data plane client security group allowed to connect to PostgreSQL"
  nullable    = false
}
