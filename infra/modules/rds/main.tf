data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-db-subnet-group"
  subnet_ids = data.aws_subnets.default.ids

  tags = {
    Service   = var.service
    Component = "rds"
    Name      = "${var.name}-db-subnet-group"
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.name}-rds"
  description = "PostgreSQL access for ${var.name}"
  vpc_id      = data.aws_vpc.default.id

  tags = {
    Service   = var.service
    Component = "rds"
    Name      = "${var.name}-rds"
  }
}

resource "aws_vpc_security_group_ingress_rule" "postgresql_from_data_plane_client" {
  security_group_id            = aws_security_group.rds.id
  referenced_security_group_id = var.data_plane_client_security_group_id
  from_port                    = var.port
  to_port                      = var.port
  ip_protocol                  = "tcp"
  description                  = "PostgreSQL from ECS data plane client tasks"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.rds.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
  description       = "Allow all outbound"
}

module "this" {
  source  = "terraform-aws-modules/rds/aws"
  version = "7.2.1"

  identifier = "${var.name}-postgres"

  engine               = var.engine
  engine_version       = var.engine_version
  family               = var.family
  major_engine_version = var.major_engine_version
  instance_class       = var.instance_class

  allocated_storage = var.allocated_storage
  storage_type      = "gp3"

  db_name  = var.db_name
  username = var.username
  port     = var.port

  manage_master_user_password = true

  create_db_subnet_group = false
  db_subnet_group_name   = aws_db_subnet_group.this.name

  vpc_security_group_ids = [aws_security_group.rds.id]

  publicly_accessible = var.publicly_accessible
  multi_az            = var.multi_az

  backup_retention_period = var.backup_retention_period
  skip_final_snapshot     = var.skip_final_snapshot
  deletion_protection     = var.deletion_protection

  storage_encrypted = var.storage_encrypted

  create_db_parameter_group = false
  create_db_option_group    = false

  tags = {
    Service   = var.service
    Component = "rds"
    Name      = "${var.name}-postgres"
  }
}
