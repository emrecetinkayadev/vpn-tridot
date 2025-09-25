variable "name" {
  type        = string
  description = "Friendly name prefix for VPC resources"
}

variable "cidr_block" {
  type        = string
  description = "CIDR block for the VPC"
}

variable "public_subnets" {
  type = map(object({
    cidr = string
    az   = string
  }))
  description = "Map of public subnet definitions"
  default     = {}
}

variable "private_subnets" {
  type = map(object({
    cidr = string
    az   = string
  }))
  description = "Map of private subnet definitions"
  default     = {}
}

variable "tags" {
  type        = map(string)
  description = "Common tags"
  default     = {}
}
