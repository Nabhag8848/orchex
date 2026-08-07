output "ecr_builder_api" {
  description = "ECR repository for the workflow builder API"
  value = {
    repository_arn         = module.ecr_builder_api.repository_arn
    repository_name        = module.ecr_builder_api.repository_name
    repository_registry_id = module.ecr_builder_api.repository_registry_id
    repository_url         = module.ecr_builder_api.repository_url
  }
}

output "alb" {
  description = "Application load balancer"
  value = {
    id                 = module.alb.id
    arn                = module.alb.arn
    arn_suffix         = module.alb.arn_suffix
    dns_name           = module.alb.dns_name
    zone_id            = module.alb.zone_id
    listeners          = module.alb.listeners
    listener_rules     = module.alb.listener_rules
    target_groups      = module.alb.target_groups
    security_group_arn = module.alb.security_group_arn
    security_group_id  = module.alb.security_group_id
    route53_records    = module.alb.route53_records
  }
}
