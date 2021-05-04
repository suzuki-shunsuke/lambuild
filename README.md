# lambuild

_Lambda => CodeBuild = lambuild_

Trigger AWS Lambda Funciton via GitHub Webhook to generate AWS CodeBuild's buildspec dynamically and start build

## Link

* [Getting Started](docs/getting-started.md)
* Configuration
  * [Lambda Function's Configuration](docs/lambda-configuration.md)
  * [lambuild.yaml](docs/lambuild-yaml.md)
  * [Secrets](docs/secret.md)
* [Expression](docs/expression.md)
* [Error Notification](docs/error-notification.md)

## Motivation

To change CodeBuild's build configuraiton dynamically by the content of event and associated pull request.
For example, running the build `test_foo` only when the service `foo` is updated in the associated pull request.

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
