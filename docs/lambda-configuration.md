# Lambda Function's Configuration

## Configuration source

`lambuild` reads a structured configuration from the configuration source.
`lambuild` supports different types of configuration sources, and we can specify the source type by the environment variable `CONFIG_SOURCE`.
The default source type is `env`.

CONFIG_SOURCE | description
--- | ---
env | The environment variable `CONFIG`
appconfig-extension | [AWS AppConfig integration with Lambda extensions](https://docs.aws.amazon.com/appconfig/latest/userguide/appconfig-integration-lambda-extensions.html)

### appconfig-extension

When we use `appconfig-extension`, we have to configure some environment variables.

name | description
--- | ---
APPCONFIG_APPLICATION_NAME | AppConfig's Application name
APPCONFIG_ENVIRONMENT_NAME | AppConfig's Environment name
APPCONFIG_CONFIG_NAME | AppConfig's Configuration name

**Note that we have to use the Lambda Function [Custom Runtime](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html) instead of [Go Runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html), because [Go Runtime doesn't support Lambda Extension](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html).**

## Configuration

e.g.

```yaml
region: us-east-1
log-level: debug
ssm-parameter:
  parameter-name:
    github-token: lambuild_github_token
    webhook-secret: lambuild_webhook_secret
build-status-context: "AWS Codebuild ({{.event.Headers.Event}})"
error-notification-template: |
  lambuild failed to procceed the request.
  Please check.
repositories:
- name: suzuki-shunsuke/test-lambuild
  hooks:
  - config: lambuild.yaml
    if: 'event.Headers.Event == "pull_request"'
  codebuild:
    project-name: test-lambuild
```

path | environment variable name | type | required | default | description
--- | --- | --- | --- | --- | ---
.region | REGION | string | true | | AWS Region
.log-level | LOG_LEVEL | string | false | info | log level of [logrus](https://github.com/sirupsen/logrus)
.ssm-parameter | | [ssm-parameter](#type-ssm-parameter) | false | | AWS Systems Manager Parameter Store configuration. Either `.ssm-parameter` or `.secrets-manager` is required
.secrets-manager | | [secrets-manager](#type-secrets-manager) | false | | AWS Secrets Manager's Secret Configuration. Either `.ssm-parameter` or `.secrets-manager` is required
.build-status-context | BUILD_STATUS_CONTEXT | [template string](#type-template-string) | false | not specified | [`build-status-config-override`'s context](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/codebuild/start-build.html)
.error-notification-template | ERROR_NOTIFICATION_TEMPLATE | [template string](#type-template-string) | false | | [Error notification template](error-notification.md)
.repositories | | [][repository](#type-repository) | true | | |

### type: ssm-parameter

path | environment variable name | type | required | description
--- | --- | --- | --- | ---
.parameter-name.github-token | SSM_PARAMETER_NAME_GITHUB_TOKEN | string | true | Systems Manager's Parameter Name against which GitHub Personal Access Token is registered
.parameter-name.webhook-secret | SSM_PARAMETER_NAME_WEBHOOK_SECRET | string | true | Systems Manager's Parameter Name against which GitHub Webhook secret is registered

### type: secrets-manager

path | environment variable name | type | required | description
--- | --- | --- | --- | ---
.secret-id | SECRETS_MANAGER_SECRET_ID | string | true | Secrets Manager's Secret ID
.version-id | SECRETS_MANAGER_VERSION_ID | string | false | Secrets Manager's Version ID

The Secret keys must be `webhook-secret` and `github-token`.

## type: repository

path | type | required | example | description
--- | --- | --- | --- | ---
.name | string | true | `suzuki-shunsuke/test-lambuild` | repository full name `<repo_owner>/<repo_name>`
.hooks | [][hook](#type-hook) | true | |
.codebuild.project-name | string | true | `test-lambuild` | 
.codebuild.assume-role-arn | string | false | | Assume Role ARN to start builds

If an event doesn't match any hook's condition, the event is ignored.

## type: hook

path | type | required | default | description
--- | --- | --- | --- | ---
.config | string | false | `lambuild.yaml` | relative path from repository's root directory to the buildspec template file or directory on the source repository. 
.if | string expression | false | "true" | the evaluated result must be a boolean. if an event doesn't match the condition, the event is ignored. If this field is empty, no event is ignored
.service-role | string | false | | CodeBuild Service Role ARN
.project-name | string | false | | CodeBuild Project Name
.assume-role-arn | string | false | | Assume Role ARN to start builds

### hook.config

If `.config` is a directory, files in the directory are treated as configuration files and procceeded in parallel, which means builds are run in parallel.
The file extension of configuration file must be `.yml` or `.yaml`, otherwise the file is ignored.

### hook.service-role

If CodeBuild Service Role has strong permissions,
dangerous code can be run in CI of pull requests.
`lambuild` supports configuring Service Role per hook,
so we can use restricted Service Role for pull requests.

e.g.

```yaml
hooks:
  - if: |
      event.Headers.Event == "push" and
      ref == "refs/heads/main"
  - if: |
      event.Headers.Event == "pull_request" and
      event.Payload.GetAction() in ["opened", "edited", "reopend", "synchronize"]
    # prevent from changing AWS resources in pull requests
    service-role: "arn:aws:iam::<AWS Account ID>:role/read-only"
```

To change Service Role, we have to add the permission `iam:PassRole` to Lambda Execution Role.

e.g.

```json
{
  "Effect": "Allow",
  "Action": "iam:PassRole",
  "Resource": "arn:aws:iam::<AWS Account ID>:role/read-only",
  "Condition": {
    "StringEquals": {"iam:PassedToService": "codebuild.amazonaws.com"},
    "StringLike": {
      "iam:AssociatedResourceARN": [
        "arn:aws:codebuild:us-east-1:<AWS Account ID>:project/test-lambuild"
      ]
    }
  }
}
```

## type: template string

`type: template string` is rendered with Go's [text/template](https://golang.org/pkg/text/template/). [sprig functions](http://masterminds.github.io/sprig/) can be used.

### build-status-context's template parameters

path | type | example | description
--- | --- | --- | ---
.event | Event | |
.pr | PullRequest | | Associated pull request
.repo | Repository | | Associated repository
.sha | string | | Associated commit SHA
.ref | string | `refs/heads/master` |

### error-notification's template parameters

path | type | example | description
--- | --- | --- | ---
.Error | Go's error | |
