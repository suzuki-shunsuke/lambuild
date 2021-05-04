# lambuild

_Lambda => CodeBuild = lambuild_

Trigger AWS Lambda Funciton via GitHub Webhook to generate AWS CodeBuild's buildspec dynamically and start build

## Link

* [Getting Started](docs/getting-started.md)
* Configuration
  * [Lambda Function's Configuration](docs/lambda-configuraion.md)
  * [Secrets](docs/secret.md)

## Motivation

To change CodeBuild's build configuraiton dynamically by the content of event and associated pull request.
For example, running the build `test_foo` only when the service `foo` is updated in the associated pull request.

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
1. Configuration file is downloaded from the source repository
1. buildspec is generated
1. Build or Batch Build is run

## LICENSE

[MIT](LICENSE)
