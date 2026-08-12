data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

module "this" {
  source  = "terraform-aws-modules/ecs/aws//modules/service"
  version = "7.5.0"

  name        = var.name
  cluster_arn = var.cluster_arn

  create_service = false

  cpu    = var.cpu
  memory = var.memory

  assign_public_ip = true
  subnet_ids       = data.aws_subnets.default.ids

  security_group_ids = [var.data_plane_client_security_group_id]

  task_exec_secret_arns = [var.database_url_secret_arn]

  container_definitions = {
    (var.container_name) = {
      essential = true
      image     = var.image

      environment = [
        { name = "GOOSE_DRIVER", value = "postgres" },
        { name = "GOOSE_MIGRATION_DIR", value = "/migrations" },
      ]

      secrets = [{
        name      = "GOOSE_DBSTRING"
        valueFrom = var.database_url_secret_arn
      }]

      readonlyRootFilesystem = false
    }
  }

  security_group_egress_rules = {
    all = {
      ip_protocol = "-1"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  tags = {
    Service   = var.service
    Component = "ecs-run-task"
    Name      = var.name
  }
}
