module "network" {
  source = "../../modules/vpc"

  name       = "prod"
  cidr_block = "10.30.0.0/16"
  tags       = var.tags

  public_subnets = {
    a = {
      cidr = "10.30.1.0/24"
      az   = "${var.aws_region}a"
    }
    b = {
      cidr = "10.30.2.0/24"
      az   = "${var.aws_region}b"
    }
  }

  private_subnets = {
    a = {
      cidr = "10.30.11.0/24"
      az   = "${var.aws_region}a"
    }
    b = {
      cidr = "10.30.12.0/24"
      az   = "${var.aws_region}b"
    }
  }
}

module "vpn_node_1" {
  source = "../../modules/vm"

  name                 = "prod-vpn-node-1"
  ami_id               = var.ami_id
  instance_type        = "t3.medium"
  subnet_id            = module.network.private_subnet_ids[0]
  vpc_id               = module.network.vpc_id
  ssh_key_name         = var.ssh_key_name
  allowed_wireguard_cidrs = ["0.0.0.0/0"]
  allowed_ssh_cidrs    = var.bastion_ssh_cidrs
  tags                 = var.tags
}

module "vpn_node_2" {
  source = "../../modules/vm"

  name                 = "prod-vpn-node-2"
  ami_id               = var.ami_id
  instance_type        = "t3.medium"
  subnet_id            = module.network.private_subnet_ids[1]
  vpc_id               = module.network.vpc_id
  ssh_key_name         = var.ssh_key_name
  allowed_wireguard_cidrs = ["0.0.0.0/0"]
  allowed_ssh_cidrs    = var.bastion_ssh_cidrs
  tags                 = var.tags
}

locals {
  prod_vpn_public_ips = compact([
    module.vpn_node_1.public_ip,
    module.vpn_node_2.public_ip,
  ])

  api_ip_addresses   = [for ip in var.api_ipv4 : trimspace(ip) if trimspace(ip) != ""]
  agent_ip_addresses = [for ip in var.agent_ipv4 : trimspace(ip) if trimspace(ip) != ""]
  panel_cname        = trimspace(var.panel_cname_target)
  status_cname       = trimspace(var.status_page_cname_target)
  ops_cname          = trimspace(var.ops_cname_target)

  records_base = [
    {
      name   = "vpn"
      type   = "A"
      values = local.prod_vpn_public_ips
      ttl    = 120
    }
  ]

  records_api = length(local.api_ip_addresses) > 0 ? [
    {
      name   = "api"
      type   = "A"
      values = local.api_ip_addresses
      ttl    = 60
    }
  ] : []

  records_agent = length(local.agent_ip_addresses) > 0 ? [
    {
      name   = "agent"
      type   = "A"
      values = local.agent_ip_addresses
      ttl    = 60
    }
  ] : []

  records_panel = local.panel_cname != "" ? [
    {
      name   = "panel"
      type   = "CNAME"
      values = [local.panel_cname]
      ttl    = 300
    }
  ] : []

  records_status = local.status_cname != "" ? [
    {
      name   = "status"
      type   = "CNAME"
      values = [local.status_cname]
      ttl    = 300
    }
  ] : []

  records_ops = local.ops_cname != "" ? [
    {
      name   = "ops"
      type   = "CNAME"
      values = [local.ops_cname]
      ttl    = 300
    }
  ] : []

  dns_records = concat(
    local.records_base,
    local.records_api,
    local.records_agent,
    local.records_panel,
    local.records_status,
    local.records_ops,
  )
}

module "dns" {
  source = "../../modules/dns"

  zone_name   = var.dns_zone_name
  create_zone = true
  tags        = var.tags

  records = local.dns_records
}
