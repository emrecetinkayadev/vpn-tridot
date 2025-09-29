variable "aws_region" {
  type        = string
  description = "AWS region"
  default     = "eu-central-1"
}

variable "tags" {
  type = map(string)
  default = {
    Project = "tridot-vpn"
    Env     = "prod"
  }
}

variable "bastion_ssh_cidrs" {
  type    = list(string)
  default = ["203.0.113.10/32"]
}

variable "ami_id" {
  type        = string
  description = "AMI ID for VPN nodes"
  default     = "ami-0abcdef1234567890"
}

variable "ssh_key_name" {
  type        = string
  description = "EC2 key pair name"
  default     = "tridot-prod"
}

variable "dns_zone_name" {
  type        = string
  description = "Primary Route53 hosted zone for production records"
  default     = "tridot.dev"
}

variable "api_ipv4" {
  type        = list(string)
  description = "IPv4 addresses to publish for api.tridot.dev"
  default     = []
}

variable "agent_ipv4" {
  type        = list(string)
  description = "IPv4 addresses to publish for agent.tridot.dev"
  default     = []
}

variable "panel_cname_target" {
  type        = string
  description = "CNAME target for panel.tridot.dev (e.g. Vercel/Netlify hostname)"
  default     = ""
}

variable "status_page_cname_target" {
  type        = string
  description = "CNAME target for status.tridot.dev"
  default     = ""
}

variable "ops_cname_target" {
  type        = string
  description = "CNAME target for ops.tridot.dev"
  default     = ""
}
