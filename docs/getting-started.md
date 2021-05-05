# Getting Started

## Requirement

* GitHub Account
* AWS Account

## Procedure

[We provide Terraform Configuration to setup AWS resources quickly for trial](/terraform).

* [Create a GitHub repository from the template repository](https://github.com/suzuki-shunsuke/example-lambuild/generate)
* [Generate GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
  * Select scope `repo`
* [Generate Webhook Secret to validate Webhook](https://docs.github.com/en/developers/webhooks-and-events/securing-your-webhooks)
* Create AWS resources
  * [Create a CodeBuild Project](https://docs.aws.amazon.com/codebuild/latest/userguide/create-project.html)
    * source: GitHub Repository
    * Enable Batch Build
    * Don't configure Webhook
  * [Create Lambda Function](https://docs.aws.amazon.com/lambda/latest/dg/getting-started-create-function.html)
  * [Create API Gateway](https://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
    * type: HTTP
  * [Create AWS Systems Manager Parameter Store's parameteres](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
    * Please see [secret](secret.md) too
  * [Configure Lambda Function's Execution Role](iam.md)
  * [Deploy Lambda Function](#deploy-lambda-function)
* [Configure GitHub Repository's webhook](https://docs.github.com/en/developers/webhooks-and-events/creating-webhooks)
  * Payload URL: AWS API Gateway Invoke URL + `/lambuild`
  * Content Type: `application/json`
  * Secret: secret you generate
  * Check `Pull Requests` and `Pushes` events
* Try running lambuild
  * Commit files and push to repository and open a pull request
  * Confirm build is run expectedly

## Sample Configuration

* [config.yaml](/terraform/config.yaml.template)
* [lambuild.yaml](/terraform/lambuild.yaml)

## Deploy Lambda Function

Download a zip file from [GitHub Releases](https://github.com/suzuki-shunsuke/lambuild/releases) and deploy it to the Lambda Function.
