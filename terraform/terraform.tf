terraform {
  required_version = ">= 0.14"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.75.1"
    }
  }
}

provider "aws" {
  region      = var.region
  max_retries = 3
}
