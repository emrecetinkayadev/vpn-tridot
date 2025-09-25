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
