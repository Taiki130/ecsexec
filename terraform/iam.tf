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

  name                               = "ecsexec"
  repo                               = "Taiki130/ecsexec"
  main_branch                        = "main"
  s3_bucket_tfmigrate_history_name   = "taikinoda-tfstate"
  s3_bucket_terraform_state_name     = "taikinoda-tfstate"
}
