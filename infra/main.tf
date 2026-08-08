# Workflow builder service
module "ecr_builder_api" {
  source = "./modules/ecr"

  repository_name = "orchex-builder-api"
}

module "alb" {
  source = "./modules/alb"

  name = "orchex-alb"
}


module "ecs" {
  source = "./modules/ecs"

  name = "orchex-cluster"
}

module "ecs_builder_api" {
  source         = "./modules/ecs_service"
  name           = "orchex-builder-api"
  cluster_arn    = module.ecs.arn
  container_name = "builder-api"
  image          = "${module.ecr_builder_api.repository_url}:latest"

  target_group_arn      = module.alb.target_groups["builder"].arn
  alb_security_group_id = module.alb.security_group_id
}
