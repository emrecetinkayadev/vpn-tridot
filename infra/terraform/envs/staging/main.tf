module "network" {
  source = "../../modules/vpc"

  name       = "staging"
  cidr_block = "10.20.0.0/16"
  tags       = var.tags

  public_subnets = {
    a = {
      cidr = "10.20.1.0/24"
      az   = "${var.aws_region}a"
    }
    b = {
      cidr = "10.20.2.0/24"
      az   = "${var.aws_region}b"
    }
  }

  private_subnets = {
    a = {
      cidr = "10.20.11.0/24"
      az   = "${var.aws_region}a"
    }
    b = {
      cidr = "10.20.12.0/24"
      az   = "${var.aws_region}b"
    }
  }
}

module "vpn_node" {
  source = "../../modules/vm"

  name                 = "staging-vpn-node-1"
  ami_id               = var.ami_id
  instance_type        = "t3.medium"
  subnet_id            = module.network.private_subnet_ids[0]
  vpc_id               = module.network.vpc_id
  ssh_key_name         = var.ssh_key_name
  allowed_wireguard_cidrs = ["0.0.0.0/0"]
  allowed_ssh_cidrs    = var.bastion_ssh_cidrs
  tags                 = var.tags
}
