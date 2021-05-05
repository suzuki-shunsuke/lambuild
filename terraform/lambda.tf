resource "aws_lambda_function" "main" {
  filename         = var.zip_path
  function_name    = var.function_name
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  source_code_hash = filebase64sha256(var.zip_path)
  # runtime          = "go1.x"
  runtime = "provided.al2"

  environment {
    variables = {
      CONFIG = templatefile("${path.module}/${var.config_path}", {
        region                            = var.region
        repo_full_name                    = var.repo_full_name
        ssm_parameter_github_token_name   = aws_ssm_parameter.github_token.name
        ssm_parameter_webhook_secret_name = aws_ssm_parameter.webhook_secret.name
        project_name                      = var.project_name
      })
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = var.lambda_role_name
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.lambda.json
}

data "aws_iam_policy_document" "lambda" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_cloudwatch_log_group" "lambda-log" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = 7
}

resource "aws_iam_role_policy" "lambda-log" {
  name   = "log"
  policy = data.aws_iam_policy_document.lambda-log.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "lambda-log" {
  statement {
    actions   = ["logs:CreateLogStream"]
    resources = ["${aws_cloudwatch_log_group.lambda-log.arn}:log-stream:*"]
  }
  statement {
    actions   = ["logs:PutLogEvents"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "read-secret" {
  name   = "read-secret"
  policy = data.aws_iam_policy_document.read-secret.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "read-secret" {
  statement {
    actions = ["ssm:GetParameter"]
    resources = [
      aws_ssm_parameter.github_token.arn,
      aws_ssm_parameter.webhook_secret.arn,
    ]
  }
}

resource "aws_iam_role_policy" "start-codebuild" {
  name   = "start-codebuild"
  policy = data.aws_iam_policy_document.start-codebuild.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "start-codebuild" {
  statement {
    actions = [
      "codebuild:StartBuildBatch",
      "codebuild:StartBuild",
    ]

    resources = [
      "*",
    ]
  }
}
