locals {
  created_zone = try(aws_route53_zone.this[0], null)
}

output "zone_id" {
  value       = local.zone_id
  description = "ID of the hosted zone used for record management."
}

output "name_servers" {
  value       = local.created_zone != null ? local.created_zone.name_servers : []
  description = "Delegation name servers when the zone is managed by this module."
}
