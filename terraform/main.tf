terraform {
  required_version = "1.9.6"
  required_providers {
    github = {
      source  = "integrations/github"
      version = "6.3.0"
    }
    sops = {
      source  = "carlpett/sops"
      version = "1.1.1"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.64.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.6"
    }
  }
  backend "s3" {
    bucket = "taikinoda-tfstate"
    key    = "ecsexec"
    region = "ap-northeast-1"
  }
}

provider "github" {
  owner = local.owner
  app_auth {
    id              = local.app_id
    installation_id = local.installation_id
    pem_file        = data.sops_file.tf_secrets.data.app_private_key
  }
}

provider "sops" {}

provider "aws" {
  region = "ap-northeast-1"
  default_tags {
    tags = {
      "Managed_by" = "${local.owner}/${local.repo}"
    }
  }
}

data "sops_file" "tf_secrets" {
  source_file = "sops/tf_secrets.yaml"
}
