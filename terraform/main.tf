terraform {
  required_version = "1.8.4"
  required_providers {
    github = {
      source  = "integrations/github"
      version = "6.2.1"
    }
    sops = {
      source  = "carlpett/sops"
      version = "1.0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.51.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.5"
    }
  }
  backend "s3" {
    bucket = "taikinoda-tfstate"
    key    = "ecsexec"
    region = "ap-northeast-1"
  }
}

provider "github" {}
provider "sops" {}

provider "aws" {
  region = "ap-northeast-1"
}
