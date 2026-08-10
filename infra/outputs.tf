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

output "ecs" {
  description = "Shared ECS Fargate cluster"
  value = {
    arn                                  = module.ecs.arn
    id                                   = module.ecs.id
    name                                 = module.ecs.name
    capacity_providers                   = module.ecs.capacity_providers
    cloudwatch_log_group_arn             = module.ecs.cloudwatch_log_group_arn
    cloudwatch_log_group_name            = module.ecs.cloudwatch_log_group_name
    cluster_capacity_providers           = module.ecs.cluster_capacity_providers
    infrastructure_iam_role_arn          = module.ecs.infrastructure_iam_role_arn
    infrastructure_iam_role_name         = module.ecs.infrastructure_iam_role_name
    infrastructure_iam_role_unique_id    = module.ecs.infrastructure_iam_role_unique_id
    node_iam_instance_profile_arn        = module.ecs.node_iam_instance_profile_arn
    node_iam_instance_profile_id         = module.ecs.node_iam_instance_profile_id
    node_iam_instance_profile_unique     = module.ecs.node_iam_instance_profile_unique
    node_iam_role_arn                    = module.ecs.node_iam_role_arn
    node_iam_role_name                   = module.ecs.node_iam_role_name
    node_iam_role_unique_id              = module.ecs.node_iam_role_unique_id
    task_exec_iam_role_arn               = module.ecs.task_exec_iam_role_arn
    task_exec_iam_role_name              = module.ecs.task_exec_iam_role_name
    task_exec_iam_role_unique_id         = module.ecs.task_exec_iam_role_unique_id
    data_plane_client_security_group_id  = module.ecs.data_plane_client_security_group_id
    data_plane_client_security_group_arn = module.ecs.data_plane_client_security_group_arn
  }
}

output "rds" {
  description = "PostgreSQL RDS instance (connection metadata only; credentials stay in Secrets Manager)"
  value = {
    db_instance_endpoint              = module.rds.db_instance_endpoint
    db_instance_address               = module.rds.db_instance_address
    db_instance_port                  = module.rds.db_instance_port
    db_instance_name                  = module.rds.db_instance_name
    db_instance_identifier            = module.rds.db_instance_identifier
    db_instance_arn                   = module.rds.db_instance_arn
    db_instance_status                = module.rds.db_instance_status
    db_instance_engine                = module.rds.db_instance_engine
    db_instance_engine_version_actual = module.rds.db_instance_engine_version_actual
    db_subnet_group_id                = module.rds.db_subnet_group_id
    security_group_id                 = module.rds.security_group_id
  }
}

output "ecs_builder_api" {
  description = "ECS service for the workflow builder API"
  value = {
    id                            = module.ecs_builder_api.id
    name                          = module.ecs_builder_api.name
    iam_role_name                 = module.ecs_builder_api.iam_role_name
    iam_role_arn                  = module.ecs_builder_api.iam_role_arn
    iam_role_unique_id            = module.ecs_builder_api.iam_role_unique_id
    container_definitions         = module.ecs_builder_api.container_definitions
    task_definition_arn           = module.ecs_builder_api.task_definition_arn
    task_definition_revision      = module.ecs_builder_api.task_definition_revision
    task_definition_family        = module.ecs_builder_api.task_definition_family
    task_exec_iam_role_name       = module.ecs_builder_api.task_exec_iam_role_name
    task_exec_iam_role_arn        = module.ecs_builder_api.task_exec_iam_role_arn
    task_exec_iam_role_unique_id  = module.ecs_builder_api.task_exec_iam_role_unique_id
    tasks_iam_role_name           = module.ecs_builder_api.tasks_iam_role_name
    tasks_iam_role_arn            = module.ecs_builder_api.tasks_iam_role_arn
    tasks_iam_role_unique_id      = module.ecs_builder_api.tasks_iam_role_unique_id
    task_set_id                   = module.ecs_builder_api.task_set_id
    task_set_arn                  = module.ecs_builder_api.task_set_arn
    task_set_stability_status     = module.ecs_builder_api.task_set_stability_status
    task_set_status               = module.ecs_builder_api.task_set_status
    autoscaling_policies          = module.ecs_builder_api.autoscaling_policies
    autoscaling_scheduled_actions = module.ecs_builder_api.autoscaling_scheduled_actions
    security_group_arn            = module.ecs_builder_api.security_group_arn
    security_group_id             = module.ecs_builder_api.security_group_id
    infrastructure_iam_role_arn   = module.ecs_builder_api.infrastructure_iam_role_arn
    infrastructure_iam_role_name  = module.ecs_builder_api.infrastructure_iam_role_name
  }
}
