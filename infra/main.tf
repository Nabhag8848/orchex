# Workflow builder service
module "ecr_builder_api" {
  source = "./modules/ecr"

  repository_name = "orchex-builder-api"
}

module "alb" {
  source = "./modules/alb"

  name = "orchex-alb"
}
