# -- inputs ---
variable "vpc" {
  description = "VPC parameters"
  type = object({
    vpc_id    = string
    subnet_id = string
  })
}

variable "key_pair" {
  description = "Key pair for EC2 instance SSH"
  type = object({
    name             = string
    private_key_file = string
  })
}

# -- providers ---
terraform {
  required_providers {
    greynoise = {
      source  = "greynoise-io/greynoise"
      version = "0.1.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.64.0"
    }
  }
}

provider "aws" {
  default_tags {
    tags = {
      Environment = "development"
      Owner       = "greynoise"
      Project     = "greynoise-tf-provider"
    }
  }

  region = "us-east-1"
}

provider "greynoise" {
  // GN_API_KEY env var is used to provide key
}

# -- main ---
locals {
  name = "greynoise-tf-provider"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}


data "aws_key_pair" "this" {
  key_name           = var.key_pair.name
  include_public_key = true
}

resource "aws_security_group" "this" {
  name        = "${local.name}-sg"
  description = "Security Group for GN sensor"

  vpc_id = var.vpc.vpc_id

  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all inbound traffic"
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
    description      = "Allow all egress traffic"
  }
}

resource "aws_instance" "this" {
  ami           = data.aws_ami.ubuntu.id
  subnet_id     = var.vpc.subnet_id
  instance_type = "t2.micro"
  key_name      = data.aws_key_pair.this.key_name
  vpc_security_group_ids = [
    aws_security_group.this.id,
  ]
}

data "greynoise_personas" "rdp" {
  search = "rdp"
  limit  = 1
}

resource "greynoise_sensor_bootstrap" "this" {
  public_ip = aws_instance.this.public_ip

  connection {
    host = aws_instance.this.public_ip
    user = "ubuntu"
    port = 22

    private_key = file(var.key_pair.private_key_file)
  }

  provisioner "remote-exec" {
    inline = [
      # ensure that script can run by waiting for cloud-init to complete
      "cloud-init status --wait > /dev/null",
      self.setup_script,
    ]
  }

  provisioner "remote-exec" {
    inline = [
      self.bootstrap_script,
    ]
    # failure is expected as SSH connection will be lost
    # once bootstrap completes and changes SSH port
    on_failure = continue
  }
}

data "greynoise_sensor" "this" {
  public_ip = aws_instance.this.public_ip
  depends_on = [
    greynoise_sensor_bootstrap.this,
  ]
}

resource "greynoise_sensor_persona" "this" {
  sensor_id  = data.greynoise_sensor.this.id
  persona_id = data.greynoise_personas.rdp.ids[0]
}

# -- outputs --
output "personas" {
  description = "RDP personas"
  value = {
    ids = data.greynoise_personas.rdp.ids
  }
}

output "sensor" {
  description = "Sensor information"
  value = {
    public_ip = aws_instance.this.public_ip
    ssh_port  = greynoise_sensor_bootstrap.this.ssh_port_selected
  }
}
