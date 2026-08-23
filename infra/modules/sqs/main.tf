data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  producer_task_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/${var.producer_task_role_name}"
  consumer_task_role_arn = var.consumer_task_role_name == null ? null : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/${var.consumer_task_role_name}"

  # Queue resource policies only allow a subset of SQS actions (*Batch is IAM-only;
  # SendMessage / DeleteMessage / ChangeMessageVisibility cover the batch APIs).
  producer_resource_policy_actions = [
    "sqs:SendMessage",
    "sqs:ReceiveMessage",
    "sqs:DeleteMessage",
    "sqs:ChangeMessageVisibility",
    "sqs:GetQueueAttributes",
    "sqs:GetQueueUrl",
  ]

  consumer_resource_policy_actions = [
    "sqs:ReceiveMessage",
    "sqs:DeleteMessage",
    "sqs:ChangeMessageVisibility",
    "sqs:GetQueueAttributes",
    "sqs:GetQueueUrl",
  ]

  # IAM task-role actions (IAM accepts *Batch).
  producer_queue_actions = concat(local.producer_resource_policy_actions, [
    "sqs:SendMessageBatch",
    "sqs:DeleteMessageBatch",
    "sqs:ChangeMessageVisibilityBatch",
  ])

  consumer_queue_actions = concat(local.consumer_resource_policy_actions, [
    "sqs:DeleteMessageBatch",
    "sqs:ChangeMessageVisibilityBatch",
  ])

  producer_principal = [{
    type        = "AWS"
    identifiers = [local.producer_task_role_arn]
  }]

  consumer_principal = local.consumer_task_role_arn == null ? [] : [{
    type        = "AWS"
    identifiers = [local.consumer_task_role_arn]
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
  queue_policy_statements = merge(
    {
      execution_api = {
        sid        = "ExecutionApiProducer"
        actions    = local.producer_resource_policy_actions
        principals = local.producer_principal
      }
    },
    var.consumer_task_role_name == null ? {} : {
      execution_worker = {
        sid        = "ExecutionWorkerConsumer"
        actions    = local.consumer_resource_policy_actions
        principals = local.consumer_principal
      }
    }
  )

  create_dlq_queue_policy = true
  dlq_queue_policy_statements = {
    execution_api = {
      sid        = "ExecutionApiDlq"
      actions    = local.producer_resource_policy_actions
      principals = local.producer_principal
    }
  }

  tags = {
    Service   = var.service
    Component = "sqs"
    Name      = var.name
  }
}
