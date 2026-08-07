output "ecr_builder_api" {
  description = "ECR repository for the workflow builder API"
  value = {
    repository_arn         = module.ecr_builder_api.repository_arn
    repository_name        = module.ecr_builder_api.repository_name
    repository_registry_id = module.ecr_builder_api.repository_registry_id
    repository_url         = module.ecr_builder_api.repository_url
  }
}
