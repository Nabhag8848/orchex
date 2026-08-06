module "ecr_api" {
  source = "./modules/ecr"

  repository_name = "orchex-api"
}
