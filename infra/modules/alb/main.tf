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
      fixed_response = {
        content_type = "application/json"
        message_body = "{\"status\":\"ok\"}"
        status_code  = "200"
      }
    }
  }

  tags = {
    Service = var.name
  }
}
