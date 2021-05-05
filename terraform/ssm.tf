resource "aws_ssm_parameter" "github_token" {
  name        = var.ssm_parameter_github_token_name
  description = "The parameter for lambuild Getting Started. GitHub Personal Access Token"
  type        = "SecureString"
  value       = var.ssm_parameter_github_token_value
}

resource "aws_ssm_parameter" "webhook_secret" {
  name        = var.ssm_parameter_webhook_secret_name
  description = "The parameter for lambuild Getting Started. GitHub Webhook Secret"
  type        = "SecureString"
  value       = var.ssm_parameter_webhook_secret_value
}
