resource "aws_secretsmanager_secret" "main" {
  name = var.secret_id
}

resource "aws_secretsmanager_secret_version" "lambuild" {
  secret_id = aws_secretsmanager_secret.main.id
  secret_string = jsonencode({
    "webhook-secret" : var.webhook_secret,
    "github-token" : var.github_token,
  })
}
