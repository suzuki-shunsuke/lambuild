# Getting Started

## Requirement

* GitHub Account
* AWS Account

## Procedure

* Create a GitHub repository for getting started
* [Create a CodeBuild Project](https://docs.aws.amazon.com/codebuild/latest/userguide/create-project.html)
  * source: GitHub Repository
  * Enable Batch Build
  * Don't configure Webhook
* [Create Lambda Function](https://docs.aws.amazon.com/lambda/latest/dg/getting-started-create-function.html)
* [Create API Gateway](https://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
  * type: HTTP
* [Generate Webhook Secret to validate Webhook](https://docs.github.com/en/developers/webhooks-and-events/securing-your-webhooks)
* [Generate GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
* [Create AWS Systems Manager Parameter Store's parameteres](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
  * Please see [secret](secret.md) too
* [Configure Lambda Function's Execution Role](iam.md)
* [Deploy Lambda Function](#deploy-lambda-function)
* [Configure GitHub Repository's webhook](https://docs.github.com/en/developers/webhooks-and-events/creating-webhooks)
* Commit files and push to repository and open a pull request
* Confirm build is run expectedly

## Sample Configuration

### Lambda Function

e.g.

```yaml
repositories:
  - name: suzuki-shunsuke/test-lambuild
    codebuild:
      project-name: test-lambuild
    hooks:
      - if: |
          event.Headers.Event == "push" and
          ref == "refs/heads/main"
      - if: |
          event.Headers.Event == "pull_request" and
          event.Payload.GetAction() in ["opened", "edited", "reopend"]
```

### lambuild.yaml

```yaml
---
version: 0.2
env:
  git-credential-helper: yes
lambuild:
  env:
    variables:
      LAMBUILD_WEBHOOK_BODY: "event.Body"
      LAMBUILD_WEBHOOK_EVENT: "event.Headers.Event"
      LAMBUILD_WEBHOOK_DELIVERY: "event.Headers.Delivery"
      LAMBUILD_HEAD_COMMIT_MSG: "getCommitMessage()"
batch:
  build-list:
    - identifier: foo
      buildspec: foo/buildspec.yaml
      if: 'any(getPRFileNames(), {# startsWith "foo/"})'
    - identifier: bar
      buildspec: buildspec/renovate.yaml
      if: 'any(getPRFileNames(), {# == "renovate.json"})'
```

## Deploy Lambda Function

Download the binary from [GitHub Releases](https://github.com/suzuki-shunsuke/lambuild/releases) and deploy it to the Lambda Function.
