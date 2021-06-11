# lambuild

[![Build Status](https://github.com/suzuki-shunsuke/lambuild/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/lambuild/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/lambuild)](https://goreportcard.com/report/github.com/suzuki-shunsuke/lambuild)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/lambuild.svg)](https://github.com/suzuki-shunsuke/lambuild)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/lambuild/master/LICENSE)

_Lambda => CodeBuild = lambuild_

`lambuild` empowers [AWS CodeBuild](https://aws.amazon.com/codebuild/) whose source provider is `GITHUB`.
Instead of [AWS CodeBuild's GitHub webhook events support](https://docs.aws.amazon.com/codebuild/latest/userguide/github-webhook.html),
`lambuild` triggers build with [GitHub Webhook](https://docs.github.com/en/developers/webhooks-and-events/webhooks), [Amazon API Gateway](https://aws.amazon.com/api-gateway/), and [AWS Lambda](https://aws.amazon.com/lambda/) to provide the following powerful features.

* [Multiple buildspec files](#multiple-buildspec-files)
* [Conditional builds](#conditional-builds)
* [Change CodeBuild Service Role conditionally](#change-codebuild-service-role-conditionally)
* [Custom Environment Variables with GitHub Webhook Event and associated Pull Request](#custom-environment-variables-with-gitHub-webhook-event-and-associated-pull-request)
* [Override Build Configuration like `image` in buildspec](#override-build-configuration-like-image-in-buildspec)
* [Run multiple builds based on the same buildspec without Batch Build](#run-multiple-builds-based-on-the-same-buildspec-without-batch-build)
* [Run Batch Build's each build conditionally](#run-batch-builds-each-build-conditionally)
* etc

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

### Multiple buildspec files

[GitHub Actions](https://docs.github.com/en/actions) supports multiple workflow files in `.github/workflows` directory.
Like GitHub Actions, `lambuild` supports multiple buildspec files.

### Conditional builds

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

### Change CodeBuild Service Role conditionally

If CodeBuild Service Role has strong permissions,
dangerous code can be run in CI of pull requests.
`lambuild` supports configuring Service Role per hook,
so we can use restricted Service Role for pull requests.

For detail, please see [here](/docs/lambda-configuration.md#hookservice-role).

### Custom Environment Variables with GitHub Webhook Event and associated Pull Request

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

### Run Batch Build's each build conditionally

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

## Architecture

```
User = (push, pull_request) => GitHub = (webhook) => API Gateway => Lambda => CodeBuild
```

![lambuild-architecture](https://user-images.githubusercontent.com/13323303/116976740-80f1d300-acfc-11eb-96f5-7fb49f0e7e25.png)

_This image is created with [diagrams.net](https://www.diagrams.net/)_

## How does it work?

1. Lambda Functions is called via GitHub Webhook
1. Request is filtered with hook configuration
1. Configuration file is downloaded from the source repository
1. buildspec is generated
1. Build or Batch Build is run

## Supported GitHub Events

* [push](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#push)
* [pull_request](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#pull_request)

## LICENSE

[MIT](LICENSE)
