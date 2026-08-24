data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  invoker_task_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/${var.invoker_task_role_name}"
}

module "this" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "8.1.2"

  function_name = var.name
  description   = "Shared Function-node JS sandbox; workers Invoke with source, input, timeout_ms"
  handler       = "index.handler"
  runtime       = "nodejs24.x"
  architectures = ["x86_64"]

  memory_size = 128
  timeout     = 300
  publish     = false

  source_path                  = "${path.module}/src"
  trigger_on_package_timestamp = false

  cloudwatch_logs_retention_in_days = 7
  attach_cloudwatch_logs_policy     = true

  recursive_loop = "Terminate"

  # publish = false → no version qualifier. Permission on $LATEST via unqualified name.
  create_current_version_allowed_triggers   = false
  create_unqualified_alias_allowed_triggers = true

  allowed_triggers = {
    ExecutionWorker = {
      principal = local.invoker_task_role_arn
      action    = "lambda:InvokeFunction"
    }
  }

  tags = {
    Service   = var.service
    Component = "lambda"
    Name      = var.name
  }
}
