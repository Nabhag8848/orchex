module "this" {
  source  = "terraform-aws-modules/ecs/aws//modules/cluster"
  version = "7.5.0"

  name = var.name

  cluster_capacity_providers = ["FARGATE"]
  default_capacity_provider_strategy = {
    FARGATE = {
      weight = 100
      base   = 1
    }
  }

  # Keep early-stage cost down; enable later if you want Container Insights.
  setting = [{
    name  = "containerInsights"
    value = "disabled"
  }]

  create_cloudwatch_log_group = false

  tags = {
    Service = var.name
  }
}
