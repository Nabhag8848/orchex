module "this" {
  source  = "terraform-aws-modules/secrets-manager/aws"
  version = "2.1.0"

  name          = var.name
  description   = var.description
  secret_string = var.secret_string

  tags = {
    Service   = var.service
    Component = "secrets-manager"
    Name      = var.name
  }
}
