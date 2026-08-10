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
    Service   = var.service
    Component = "ecs"
    Name      = var.name
  }
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "data_plane_client" {
  name        = "${var.name}-data-plane-client"
  description = "Shared client security group for ECS tasks that access the data plane"
  vpc_id      = data.aws_vpc.default.id

  tags = {
    Service   = var.service
    Component = "ecs"
    Name      = "${var.name}-data-plane-client"
  }
}

