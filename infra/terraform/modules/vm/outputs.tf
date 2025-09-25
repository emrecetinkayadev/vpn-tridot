output "instance_id" {
  value = aws_instance.this.id
}

output "security_group_id" {
  value = aws_security_group.this.id
}

output "public_ip" {
  value       = try(aws_eip.this[0].public_ip, null)
  description = "Elastic IP address when allocated"
}
