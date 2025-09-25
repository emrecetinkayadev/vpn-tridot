variable "name" {
  type        = string
  description = "Instance name"
}

variable "ami_id" {
  type        = string
  description = "AMI ID to launch"
}

variable "instance_type" {
  type        = string
  description = "AWS instance type"
  default     = "t3.small"
}

variable "subnet_id" {
  type        = string
  description = "Subnet ID for instance"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "wireguard_port" {
  type        = number
  description = "UDP port for WireGuard"
  default     = 51820
}

variable "allowed_wireguard_cidrs" {
  type        = list(string)
  description = "List of CIDRs allowed to reach WireGuard"
  default     = ["0.0.0.0/0"]
}

variable "allowed_ssh_cidrs" {
  type        = list(string)
  description = "List of CIDRs allowed for SSH"
  default     = []
}

variable "ssh_key_name" {
  type        = string
  description = "EC2 key pair name"
  default     = null
}

variable "root_volume_size" {
  type        = number
  description = "Root disk size"
  default     = 40
}

variable "allocate_eip" {
  type        = bool
  description = "Whether to allocate an Elastic IP"
  default     = true
}

variable "tags" {
  type        = map(string)
  description = "Common tags"
  default     = {}
}
