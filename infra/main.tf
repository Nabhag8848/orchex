# Workflow builder service
module "ecr_builder_api" {
  source = "./modules/ecr"

  repository_name = "orchex-builder-api"
  service         = "builder-api"
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
  service        = "builder-api"
  cluster_arn    = module.ecs.arn
  container_name = "builder-api"
  image          = "${module.ecr_builder_api.repository_url}:latest"

  target_group_arn                    = module.alb.target_groups["builder"].arn
  alb_security_group_id               = module.alb.security_group_id
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}

module "rds" {
  source = "./modules/rds"

  name                                = "orchex"
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}
