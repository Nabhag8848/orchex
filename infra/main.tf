# Workflow builder service
module "ecr_builder_api" {
  source = "./modules/ecr"

  repository_name = "orchex-builder-api"
  service         = "builder-api"
}

module "ecr_execution_api" {
  source = "./modules/ecr"

  repository_name = "orchex-execution-api"
  service         = "execution-api"
}

module "ecr_db_migrate" {
  source = "./modules/ecr"

  repository_name = "orchex-db-migrate"
  service         = "shared"
}

module "alb" {
  source = "./modules/alb"

  name = "orchex-alb"
}

module "sqs_node_jobs" {
  source = "./modules/sqs"

  name                   = "orchex-node-jobs"
  allowed_task_role_name = "orchex-execution-api"
}


module "ecs" {
  source = "./modules/ecs"

  name = "orchex-cluster"
}

module "rds" {
  source = "./modules/rds"

  name                                = "orchex"
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}

data "aws_secretsmanager_secret_version" "rds_master" {
  secret_id = module.rds.db_instance_master_user_secret_arn

  depends_on = [module.rds]
}

locals {
  rds_master = jsondecode(data.aws_secretsmanager_secret_version.rds_master.secret_string)

  database_url = format(
    "postgres://%s:%s@%s:%s/%s?sslmode=require",
    urlencode(local.rds_master.username),
    urlencode(local.rds_master.password),
    module.rds.db_instance_address,
    module.rds.db_instance_port,
    module.rds.db_instance_name,
  )
}

module "database_url" {
  source = "./modules/secrets_manager"

  name          = "orchex/DATABASE_URL"
  description   = "Shared Postgres connection URL for ECS services"
  secret_string = local.database_url
  service       = "shared"
}

module "ecs_db_migrate" {
  source         = "./modules/ecs_run_task"
  name           = "orchex-db-migrate"
  cluster_arn    = module.ecs.arn
  container_name = "migrate"
  image          = "${module.ecr_db_migrate.repository_url}:latest"

  database_url_secret_arn             = module.database_url.arn
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}

module "ecs_builder_api" {
  source         = "./modules/ecs_service"
  name           = "orchex-builder-api"
  service        = "builder-api"
  cluster_arn    = module.ecs.arn
  container_name = "builder-api"
  image          = "${module.ecr_builder_api.repository_url}:latest"

  target_group_arn                    = module.alb.target_groups["builder"].arn
  alb_security_group_id               = module.alb.security_group_id
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
  database_url_secret_arn             = module.database_url.arn
}

module "ecs_execution_api" {
  source         = "./modules/ecs_service"
  name           = "orchex-execution-api"
  service        = "execution-api"
  cluster_arn    = module.ecs.arn
  container_name = "execution-api"
  image          = "${module.ecr_execution_api.repository_url}:latest"

  target_group_arn                    = module.alb.target_groups["execution"].arn
  alb_security_group_id               = module.alb.security_group_id
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
  database_url_secret_arn             = module.database_url.arn

  # Stable task role name so the SQS resource policy can allow this principal
  # without a Terraform cycle (queue ARN ↔ role ARN).
  tasks_iam_role_name            = "orchex-execution-api"
  tasks_iam_role_use_name_prefix = false
  extra_environment = [
    { name = "SQS_QUEUE_URL", value = module.sqs_node_jobs.queue_url },
    { name = "SQS_DLQ_URL", value = module.sqs_node_jobs.dead_letter_queue_url },
  ]
  tasks_iam_role_statements = [{
    sid     = "NodeJobsQueue"
    actions = module.sqs_node_jobs.queue_actions
    resources = [
      module.sqs_node_jobs.queue_arn,
      module.sqs_node_jobs.dead_letter_queue_arn,
    ]
  }]
}
