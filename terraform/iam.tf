data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]
  client_id_list  = ["sts.amazonaws.com"]
}

module "aws" {
  source = "github.com/suzuki-shunsuke/terraform-aws-tfaction?ref=v0.2.1"

  name                             = "ecsexec"
  repo                             = "Taiki130/ecsexec"
  main_branch                      = "main"
  s3_bucket_tfmigrate_history_name = "taikinoda-tfstate"
  s3_bucket_terraform_state_name   = "taikinoda-tfstate"
}

data "aws_iam_role" "main" {
  for_each = local.gha_iam_roles

  name = "GitHubActions_Terraform_ecsexec_${each.key}"
}

data "aws_iam_policy" "admin" {
  name = "AdministratorAccess"
}

resource "aws_iam_role_policy_attachment" "main" {
  for_each = local.gha_iam_roles

  role       = data.aws_iam_role.main[each.key].name
  policy_arn = data.aws_iam_policy.admin.arn
}

locals {
  gha_iam_roles = [
    "terraform_apply",
    "terraform_plan",
    "tfmigrate_apply",
    "tfmigrate_plan",
  ]
}
