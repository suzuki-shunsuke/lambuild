# lambuild

AWS Lambda Function to generate AWS CodeBuild's buildspec dynamically and trigger Batch Build via GitHub Webhook

## Why `lambuild` is needed?

To change CodeBuild's buildspec dynamically according to GitHub Webhook event.

## Supported GitHub Event

* push
* pull_request

## Archtecture

```
GitHub Webhook => API Gateway => Lambda => CodeBuild
```

## How to work

1. Lambda Functions is called via GitHub Webhook
1. Request is filtered with hook configuration
1. Configuration file is downloaded from the build repository
1. buildspec is generated
1. Build or Batch Build is run

## Setup

* Create AWS Resources
  * HTTP API Gateway
  * Lambda Function
  * Systems Manager Parameter Store
  * CodeBuild Project
* Create GitHub Repository's Webhook
* Add `lambuild.yaml` to build repositories

## Configuration

### Lambda Function's Environment Variables

#### Required

* CONFIG
* REGION
* SSM_PARAMETER_NAME_GITHUB_TOKEN
* SSM_PARAMETER_NAME_WEBHOOK_SECRET

#### Optional

* LOG_LEVEL
* BUILD_STATUS_CONTEXT

### AWS Systems Manager Parameter Store

* GitHub Access Token (`SSM_PARAMETER_NAME_GITHUB_TOKEN`)
* GitHub Webhook Secret (`SSM_PARAMETER_NAME_WEBHOOK_SECRET`)

### CONFIG

e.g.

```yaml
---
repositories:
  - name: suzuki-shunsuke/test-lambuild
    codebuild:
      project-name: test-lambuild
    hooks:
      - event:
          - push
        ref: |
          equal refs/heads/master
          equal refs/heads/feat/first-pr
        config: lambuild.yaml
      - event:
          - pull_request
        base-ref: |
          equal master
        author: |
          equal renovate[bot]
        label: |
          equal terraform-provider
        config: lambuild.yaml
```

### lambuild.yaml

lambuild.yaml is the extention of buildspec.
`lambuild` field is added.

e.g.

```yaml
---
version: 0.2
env:
  git-credential-helper: yes
batch:
  build-graph:
    - identifier: validate_renovate
      buildspec: buildspec/validate-renovate.yaml
      lambuild:
        filter:
          - event:
              - push
```

### Built in environment variables

The following environment variables are added to buildspec.

* LAMBUILD_WEBHOOK_BODY
* LAMBUILD_WEBHOOK_EVENT
* LAMBUILD_WEBHOOK_DELIVERY
* LAMBUILD_HEAD_COMMIT_MSG

### Matchfile filter

[matchfile-parser](https://github.com/suzuki-shunsuke/matchfile-parser) is used to filter requests and Batch Build's build.

## LICENSE

[MIT](LICENSE)
