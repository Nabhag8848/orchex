output "lambda_function_arn" {
  description = "ARN of the sandbox function"
  value       = module.this.lambda_function_arn
}

output "lambda_function_name" {
  description = "Name of the sandbox function"
  value       = module.this.lambda_function_name
}

output "lambda_function_invoke_arn" {
  description = "Invoke ARN of the sandbox function"
  value       = module.this.lambda_function_invoke_arn
}

output "lambda_role_arn" {
  description = "Execution role ARN (logs only)"
  value       = module.this.lambda_role_arn
}

output "lambda_role_name" {
  description = "Execution role name"
  value       = module.this.lambda_role_name
}

output "lambda_cloudwatch_log_group_name" {
  description = "CloudWatch log group name"
  value       = module.this.lambda_cloudwatch_log_group_name
}

output "lambda_cloudwatch_log_group_arn" {
  description = "CloudWatch log group ARN"
  value       = module.this.lambda_cloudwatch_log_group_arn
}

output "invoker_task_role_arn" {
  description = "IAM role ARN allowed to Invoke (worker)"
  value       = local.invoker_task_role_arn
}

output "invoke_actions" {
  description = "IAM actions for the worker task role"
  value       = ["lambda:InvokeFunction"]
}
