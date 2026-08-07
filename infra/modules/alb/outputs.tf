output "id" {
  description = "The ID and ARN of the load balancer we created"
  value       = module.this.id
}

output "arn" {
  description = "The ID and ARN of the load balancer we created"
  value       = module.this.arn
}

output "arn_suffix" {
  description = "ARN suffix of our load balancer - can be used with CloudWatch"
  value       = module.this.arn_suffix
}

output "dns_name" {
  description = "The DNS name of the load balancer"
  value       = module.this.dns_name
}

output "zone_id" {
  description = "The zone_id of the load balancer to assist with creating DNS records"
  value       = module.this.zone_id
}

output "listeners" {
  description = "Map of listeners created and their attributes"
  value       = module.this.listeners
}

output "listener_rules" {
  description = "Map of listeners rules created and their attributes"
  value       = module.this.listener_rules
}

output "target_groups" {
  description = "Map of target groups created and their attributes"
  value       = module.this.target_groups
}

output "security_group_arn" {
  description = "Amazon Resource Name (ARN) of the security group"
  value       = module.this.security_group_arn
}

output "security_group_id" {
  description = "ID of the security group"
  value       = module.this.security_group_id
}

output "route53_records" {
  description = "The Route53 records created and attached to the load balancer"
  value       = module.this.route53_records
}
