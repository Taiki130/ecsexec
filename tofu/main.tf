terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "6.0.0"
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

  repository      = "ecsexec"
  secret_name     = each.key
  plaintext_value = each.value
}

data "sops_file" "secrets" {
  source_file = "secrets.yaml"
}
