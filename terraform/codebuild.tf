resource "aws_codebuild_project" "main" {
  name         = var.project_name
  service_role = aws_iam_role.codebuild.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/amazonlinux2-x86_64-standard:3.0"
    image_pull_credentials_type = "CODEBUILD"
    type                        = "LINUX_CONTAINER"
  }

  source {
    location            = "https://github.com/${var.repo_full_name}.git"
    report_build_status = true
    type                = "GITHUB"
  }

  build_batch_config {
    service_role    = aws_iam_role.codebuild.arn
    timeout_in_mins = 60
    restrictions {
      compute_types_allowed  = []
      maximum_builds_allowed = 100
    }
  }
}

resource "aws_iam_role" "codebuild" {
  assume_role_policy = data.aws_iam_policy_document.assume-role-codebuild.json
  name               = var.codebuild_role_name
  path               = "/service-role/"
}

data "aws_iam_policy_document" "assume-role-codebuild" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["codebuild.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "codebuild" {
  name   = "codebuild"
  policy = data.aws_iam_policy_document.codebuild.json
  role   = aws_iam_role.codebuild.name
}

data "aws_iam_policy_document" "codebuild" {
  statement {
    actions = [
      "codebuild:*",
    ]

    resources = [
      aws_codebuild_project.main.arn,
    ]
  }
}

resource "aws_cloudwatch_log_group" "codebuild" {
  name              = "/aws/codebuild/${var.project_name}"
  retention_in_days = 7
}

resource "aws_iam_role_policy" "codebuild-log" {
  name   = "log"
  policy = data.aws_iam_policy_document.codebuild-log.json
  role   = aws_iam_role.codebuild.name
}

data "aws_iam_policy_document" "codebuild-log" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "${aws_cloudwatch_log_group.codebuild.arn}:*",
      "${aws_cloudwatch_log_group.codebuild.arn}:*:*",
    ]
  }
}
