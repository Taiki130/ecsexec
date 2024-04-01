terraform {
  required_version = "1.6.2" # renovate: datasource=github-releases depName=opentofu/opentofu
  required_providers {
    github = {
      source  = "integrations/github"
      version = "6.2.1"
    }
    sops = {
      source  = "carlpett/sops"
      version = "1.0.0"
    }
  }
}

provider "github" {}
provider "sops" {}

resource "github_actions_secret" "secrets" {
  for_each = nonsensitive(data.sops_file.secrets.data)

  repository      = local.repo
  secret_name     = each.key
  plaintext_value = each.value
}

data "sops_file" "secrets" {
  source_file = "secrets.yaml"
}

locals {
  repo = "ecsexec"
}
