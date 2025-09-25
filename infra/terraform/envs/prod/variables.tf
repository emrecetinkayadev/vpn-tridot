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
