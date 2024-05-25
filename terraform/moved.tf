moved {
  from = aws_iam_role.main
  to   = aws_iam_role.terraform
}

moved {
  from = aws_iam_role_policy_attachment.main
  to   = aws_iam_role_policy_attachment.terraform_admin
}
