output "ecr_api" {
  description = "ECR repository details for the API service"
  value = {
    repository_arn         = module.ecr_api.repository_arn
    repository_name        = module.ecr_api.repository_name
    repository_registry_id = module.ecr_api.repository_registry_id
    repository_url         = module.ecr_api.repository_url
  }
}
