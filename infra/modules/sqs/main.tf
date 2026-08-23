data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  allowed_task_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/${var.allowed_task_role_name}"

  # Identity + resource policies share this set. ChangeMessageVisibility is
  # required for per-message leases (ReceiveMessage VisibilityTimeout is not enough
  # when a node timeout exceeds the queue default).
  queue_actions = [
    "sqs:SendMessage",
    "sqs:SendMessageBatch",
    "sqs:ReceiveMessage",
    "sqs:DeleteMessage",
    "sqs:DeleteMessageBatch",
    "sqs:ChangeMessageVisibility",
    "sqs:ChangeMessageVisibilityBatch",
    "sqs:GetQueueAttributes",
    "sqs:GetQueueUrl",
  ]

  execution_api_principal = [{
    type        = "AWS"
    identifiers = [local.allowed_task_role_arn]
  }]
}

module "this" {
  source  = "terraform-aws-modules/sqs/aws"
  version = "5.2.2"

  name = var.name

  fifo_queue                 = false
  sqs_managed_sse_enabled    = true
  message_retention_seconds  = var.message_retention_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  create_dlq                      = true
  dlq_name                        = var.dlq_name
  dlq_sqs_managed_sse_enabled     = true
  dlq_message_retention_seconds   = var.dlq_message_retention_seconds
  create_dlq_redrive_allow_policy = true

  redrive_policy = {
    maxReceiveCount = var.max_receive_count
  }

  create_queue_policy = true
  queue_policy_statements = {
    execution_api = {
      sid        = "ExecutionApiOnly"
      actions    = local.queue_actions
      principals = local.execution_api_principal
    }
  }

  create_dlq_queue_policy = true
  dlq_queue_policy_statements = {
    execution_api = {
      sid        = "ExecutionApiOnly"
      actions    = local.queue_actions
      principals = local.execution_api_principal
    }
  }

  tags = {
    Service   = var.service
    Component = "sqs"
    Name      = var.name
  }
}
