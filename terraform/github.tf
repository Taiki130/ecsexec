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
  owner           = "Taiki130"
  repo            = "ecsexec"
  app_id          = 905964
  installation_id = 51149317
}
