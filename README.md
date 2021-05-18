# lambuild

[![Build Status](https://github.com/suzuki-shunsuke/lambuild/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/lambuild/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/lambuild)](https://goreportcard.com/report/github.com/suzuki-shunsuke/lambuild)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/lambuild.svg)](https://github.com/suzuki-shunsuke/lambuild)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/lambuild/master/LICENSE)

_Lambda => CodeBuild = lambuild_

Extend AWS CodeBuild with Lambda.

## Link

* [Getting Started](docs/getting-started.md)
* Configuration
  * [Lambda Function's Configuration](docs/lambda-configuration.md)
  * [lambuild.yaml](docs/lambuild-yaml.md)
  * [Secrets](docs/secret.md)
  * [Sample Terraform Configuration](terraform)
* [Expression](docs/expression.md)
* [Error Notification](docs/error-notification.md)
* [Practice](docs/practice.md)

## Feature

### Run Batch Build's each build conditionally.

e.g.

```yaml
version: 0.2
batch:
  build-list:
    - identifier: foo
      buildspec: foo/buildspec.yaml
      if: 'any(getPRFileNames(), {# startsWith "foo/"})'
    - identifier: renovate
      buildspec: buildspec/renovate.yaml
      if: 'any(getPRFileNames(), {# == "renovate.json"})'
```

### Support multiple buildspec files

[GitHub Actions](https://docs.github.com/en/actions) supports multiple workflow files on `.github/workflows` directory.
Like GitHub Actions, `lambuild` supports multiple buildspec files.

### Run builds conditionally.

e.g.

```yaml
version: 0.2
lambuild:
  if: 'event.Headers.Event == "pull_request"'
  build-status-context: "foo ({{.event.Headers.Event}})"
phases:
  build:
    commands:
      - "echo foo"
```

### Support to define custom Environment Variables with GitHub Webhook Event and associated Pull Request

e.g.

```yaml
version: 0.2
lambuild:
  env:
    variables:
      LAMBUILD_WEBHOOK_BODY: "event.Body"
      LAMBUILD_WEBHOOK_EVENT: "event.Headers.Event"
      LAMBUILD_WEBHOOK_DELIVERY: "event.Headers.Delivery"
      LAMBUILD_HEAD_COMMIT_MSG: "getCommitMessage()"
phases:
  build:
    commands:
      - 'echo $LAMBUILD_HEAD_COMMIT_MSG'
```

### Override Build Configuration like `image` in buildspec

e.g.

```yaml
version: 0.2
lambuild:
  build-status-context: "foo ({{.event.Headers.Event}})"
  image: aws/codebuild/standard:5.0
  compute-type: BUILD_GENERAL1_SMALL
phases:
  build:
    commands:
      - "echo foo"
```

### Run multiple builds based on the same buildspec without Batch Build

e.g.

```yaml
version: 0.2
lambuild:
  env:
    variables:
      NAME: "item.name"
  items:
  - image: aws/codebuild/standard:5.0
    param:
      name: foo
  - param:
      name: bar
phases:
  build:
    commands:
      - "echo NAME: $NAME"
```

Maybe you prefer this feature rather than Batch Build, because

* It takes time to run Batch Build
* Batch Build is a little inconvenient

## Architecture

```
User = (push, pull_request) => GitHub = (webhook) => API Gateway => Lambda => CodeBuild
```

![lambuild-architecture](https://user-images.githubusercontent.com/13323303/116976740-80f1d300-acfc-11eb-96f5-7fb49f0e7e25.png)

_This image is created with [diagrams.net](https://www.diagrams.net/)_

## How to work

1. Lambda Functions is called via GitHub Webhook
1. Request is filtered with hook configuration
1. Configuration file is downloaded from the source repository
1. buildspec is generated
1. Build or Batch Build is run

## Supported GitHub Event

* [push](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#push)
* [pull_request](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#pull_request)

## LICENSE

[MIT](LICENSE)
