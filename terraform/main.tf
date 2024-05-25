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
      version = "5.51.1"
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

provider "github" {
  owner = "Taiki130"
  app_auth {
    id              = 905964
    installation_id = 51149317
    pem_file        = data.sops_file.tf_secrets.data.app_private_key
  }
}

provider "sops" {}

provider "aws" {
  region = "ap-northeast-1"
}

data "sops_file" "tf_secrets" {
  source_file = "tf_secrets.yaml"
}
