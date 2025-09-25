variable "aws_region" {
  type        = string
  description = "AWS region"
  default     = "eu-central-1"
}

variable "tags" {
  type        = map(string)
  description = "Common tags"
  default = {
    Project = "tridot-vpn"
    Env     = "staging"
  }
}

variable "bastion_ssh_cidrs" {
  type        = list(string)
  description = "CIDRs allowed to SSH into nodes"
  default     = ["198.51.100.10/32"]
}

variable "ami_id" {
  type        = string
  description = "AMI ID for VPN nodes"
  default     = "ami-0abcdef1234567890"
}

variable "ssh_key_name" {
  type        = string
  description = "EC2 key pair"
  default     = "tridot-staging"
}
