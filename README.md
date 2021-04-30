# lambuild

AWS Lambda Function to generate AWS CodeBuild's buildspec dynamically and trigger Batch Build via GitHub Webhook

## Why `lambuild` is needed?

To change CodeBuild's buildspec dynamically according to GitHub Webhook event.
And to trigger builds via GitHub events like change pull request labels.

## Archtecture

```
GitHub Webhook => API Gateway => Lambda => CodeBuild
```

## Setup

* Create AWS Resources
  * HTTP API Gateway
  * Lambda Function
  * Systems Manager Parameter Store
  * S3 Bucket
  * CodeBuild Project
* Create GitHub Repository's Webhook
* Add `lambuild.yaml` to target repositories

## Configuration

### Lambda Function's Environment Variables

* CONFIG
* REGION
* BUILDSPEC_S3_BUCKET
* BUILDSPEC_S3_KEY_PREFIX
* SSM_PARAMETER_NAME_GITHUB_TOKEN
* SSM_PARAMETER_NAME_WEBHOOK_SECRET

### AWS Systems Manager Parameter Store

* GitHub Access Token (`SSM_PARAMETER_NAME_GITHUB_TOKEN`)
* GitHub Webhook Secret (`SSM_PARAMETER_NAME_WEBHOOK_SECRET`)

### CONFIG

```yaml
---
repositories:
  - name: suzuki-shunsuke/test-lambuild
    codebuild:
      project_name: test-lambuild
    hooks:
      - event: push
        refs: |
          equal refs/heads/master
          equal refs/heads/feat/first-pr
        config: lambuild.yaml
```

## LICENSE

[MIT](LICENSE)
