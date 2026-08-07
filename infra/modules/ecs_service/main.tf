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

  cpu    = var.cpu
  memory = var.memory

  desired_count = var.desired_count

  enable_autoscaling = false

  assign_public_ip = true
  subnet_ids       = data.aws_subnets.default.ids

  container_definitions = {
    (var.container_name) = {
      essential = true
      image     = var.image
      portMappings = [{
        name          = var.container_name
        containerPort = var.container_port
        protocol      = "tcp"
      }]

      # Distroless image; keep filesystem read-only.
      readonlyRootFilesystem = true
    }
  }

  # Temporary: open the app port so you can hit the task public IP.
  # Later: restrict to the ALB security group only.
  security_group_ingress_rules = {
    app = {
      from_port   = var.container_port
      to_port     = var.container_port
      ip_protocol = "tcp"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  security_group_egress_rules = {
    all = {
      ip_protocol = "-1"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  tags = {
    Service = var.name
  }
}
