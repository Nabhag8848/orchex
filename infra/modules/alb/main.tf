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
  source  = "terraform-aws-modules/alb/aws"
  version = "10.5.0"

  name    = var.name
  vpc_id  = data.aws_vpc.default.id
  subnets = data.aws_subnets.default.ids

  enable_deletion_protection = false

  security_group_ingress_rules = {
    http = {
      from_port   = 80
      to_port     = 80
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

  listeners = {
    http = {
      port     = 80
      protocol = "HTTP"
      forward = {
        target_group_key = "builder"
      }
    }
  }

  target_groups = {
    builder = {
      name              = "orchex-builder-api"
      protocol          = "HTTP"
      port              = 8080
      target_type       = "ip"
      create_attachment = false
      health_check = {
        enabled  = true
        path     = "/health/builder"
        protocol = "HTTP"
        matcher  = "200"
      }
    }
  }

  tags = {
    Service   = var.service
    Component = "alb"
    Name      = var.name
  }
}
