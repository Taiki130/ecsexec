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
