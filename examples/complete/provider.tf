terraform {
  required_providers {
    greynoise = {
      source  = "GreyNoise-Intelligence/greynoise"
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
