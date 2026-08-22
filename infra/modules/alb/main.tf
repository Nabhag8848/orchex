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
        message_body = "{\"error\":\"not found\"}"
        status_code  = "404"
      }

      rules = {
        builder = {
          priority = 10
          actions = [{
            forward = {
              target_group_key = "builder"
            }
          }]
          conditions = [{
            path_pattern = {
              values = [
                "/health/builder",
                "/health/builder/*",
                "/v1/workflows",
                "/v1/workflows/*",
              ]
            }
          }]
        }

        execution = {
          priority = 20
          actions = [{
            forward = {
              target_group_key = "execution"
            }
          }]
          conditions = [{
            path_pattern = {
              values = [
                "/health/execution",
                "/health/execution/*",
                "/v1/runs",
                "/v1/runs/*",
              ]
            }
          }]
        }
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
    execution = {
      name              = "orchex-execution-api"
      protocol          = "HTTP"
      port              = 8080
      target_type       = "ip"
      create_attachment = false
      health_check = {
        enabled  = true
        path     = "/health/execution"
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
