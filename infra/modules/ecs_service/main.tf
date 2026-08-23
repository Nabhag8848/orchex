data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

locals {
  container_environment = concat(
    [
      { name = "HTTP_ADDR", value = ":8080" },
    ],
    var.extra_environment,
  )

  container_secrets = var.database_url_secret_arn != null ? [
    {
      name      = "DATABASE_URL"
      valueFrom = var.database_url_secret_arn
    },
  ] : []
}

module "this" {
  source  = "terraform-aws-modules/ecs/aws//modules/service"
  version = "7.5.0"

  name        = var.name
  cluster_arn = var.cluster_arn

  cpu    = var.cpu
  memory = var.memory

  desired_count = var.desired_count

  enable_autoscaling = false

  assign_public_ip = true
  subnet_ids       = data.aws_subnets.default.ids

  security_group_ids = [var.data_plane_client_security_group_id]

  task_exec_secret_arns = var.database_url_secret_arn != null ? [var.database_url_secret_arn] : []

  tasks_iam_role_name            = var.tasks_iam_role_name
  tasks_iam_role_use_name_prefix = var.tasks_iam_role_use_name_prefix
  tasks_iam_role_statements      = var.tasks_iam_role_statements

  container_definitions = {
    (var.container_name) = {
      essential = true
      image     = var.image
      portMappings = [{
        name          = var.container_name
        containerPort = var.container_port
        protocol      = "tcp"
      }]

      environment = local.container_environment
      secrets     = local.container_secrets

      # Distroless image; keep filesystem read-only.
      readonlyRootFilesystem = true
    }
  }

  load_balancer = var.attach_load_balancer ? {
    alb = {
      target_group_arn = var.target_group_arn
      container_name   = var.container_name
      container_port   = var.container_port
    }
  } : {}

  security_group_ingress_rules = var.attach_load_balancer ? {
    app = {
      from_port                    = var.container_port
      to_port                      = var.container_port
      ip_protocol                  = "tcp"
      referenced_security_group_id = var.alb_security_group_id
    }
  } : {}

  security_group_egress_rules = {
    all = {
      ip_protocol = "-1"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  tags = {
    Service   = var.service
    Component = "ecs-service"
    Name      = var.name
  }
}
