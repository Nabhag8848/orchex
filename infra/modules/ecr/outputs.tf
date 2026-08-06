output "repository_arn" {
  description = "Full ARN of the repository"
  value       = module.this.repository_arn
}

output "repository_name" {
  description = "Name of the repository"
  value       = module.this.repository_name
}

output "repository_registry_id" {
  description = "The registry ID where the repository was created"
  value       = module.this.repository_registry_id
}

output "repository_url" {
  description = "The URL of the repository"
  value       = module.this.repository_url
}