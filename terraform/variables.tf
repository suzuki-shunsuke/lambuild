variable "region" {
  type = string
}

variable "zip_path" {
  type        = string
  description = ""
  default     = "lambuild_linux_amd64.zip"
}

variable "config_path" {
  type        = string
  description = ""
  default     = "config.yaml"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "test-lambuild"
}

variable "lambda_role_name" {
  type        = string
  description = ""
  default     = "test-lambuild"
}

variable "repo_full_name" {
  type        = string
  description = "source repository full name"
}

variable "codebuild_role_name" {
  type        = string
  description = "IAM Role Name for CodeBuild"
  default     = "test-lambuild-codebuild"
}

variable "api_gateway_name" {
  type        = string
  description = "API Gateway name"
  default     = "test-lambuild"
}

variable "project_name" {
  type        = string
  description = "CodeBuild Project name"
  default     = "test-lambuild"
}

variable "ssm_parameter_github_token_name" {
  type    = string
  default = "/test-lambuild/github_token"
}

variable "ssm_parameter_github_token_value" {
  type      = string
  sensitive = true
}

variable "ssm_parameter_webhook_secret_name" {
  type    = string
  default = "/test-lambuild/webhook_secret"
}

variable "ssm_parameter_webhook_secret_value" {
  type      = string
  sensitive = true
}
