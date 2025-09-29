variable "zone_name" {
  type        = string
  description = "Fully-qualified domain name for the Route53 zone (without trailing dot)."
}

variable "create_zone" {
  type        = bool
  description = "Whether this module should create the Route53 zone."
  default     = false
}

variable "existing_zone_id" {
  type        = string
  description = "Existing Route53 zone ID to reuse when create_zone is false."
  default     = null
}

variable "comment" {
  type        = string
  description = "Optional comment for the hosted zone."
  default     = ""
}

variable "force_destroy" {
  type        = bool
  description = "Allow Terraform to delete the zone even if it contains records."
  default     = false
}

variable "tags" {
  type        = map(string)
  description = "Common tags applied to the hosted zone."
  default     = {}
}

variable "default_ttl" {
  type        = number
  description = "Default TTL applied to records when not provided explicitly."
  default     = 60
}

variable "records" {
  type = list(object({
    name            = string
    type            = string
    values          = optional(list(string), [])
    ttl             = optional(number)
    allow_overwrite = optional(bool)
  }))
  description = "List of DNS records to manage within the zone."
  default     = []
}
