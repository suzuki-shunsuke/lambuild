# Getting Started

## Requirement

* GitHub Account
* AWS Account

## Prepare

* Create a repository for getting started
* Create a CodeBuild Project
  * source: GitHub Repository
  * Enable Batch Build
  * Don't configure Webhook
* Create Lambda Function
* Create API Gateway
  * type: HTTP
* Generate Webhook Secret to validate Webhook
* Generate GitHub Personal Access Token
* Create AWS Systems Manager Parameter Store's parameteres
* Configure Lambda Function's Execution Role
* Deploy Lambda Function
* Configure GitHub Repository's webhook
* Commit files and push to repository and open a pull request
* Confirm build is started expectedly

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
