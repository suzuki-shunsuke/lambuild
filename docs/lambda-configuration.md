# Lambda Function's Configuration

## Required Environment Variables

name | example | description
--- | --- | ---
CONFIG | | YAML string to filter events and map it to repository configuration and CodeBuild Project
REGION | `us-east-1` | AWS Region where Systems Manamager's Parameter Store and CodeBuild Project exist
SSM_PARAMETER_NAME_GITHUB_TOKEN | `lambuild_github_token` | Systems Manager's Parameter Name which GitHub Personal Access Token is registered
SSM_PARAMETER_NAME_WEBHOOK_SECRET | `lambuild_webhook_secret` | Systems Manager's Parameter Name which GitHub Webhook secret is registered

## CONFIG

e.g.

```yaml
repositories:
- name: suzuki-shunsuke/test-lambuild
  hooks:
  - config: lambuild.yaml
    if: 'event.Headers.Event == "pull_request"'
  codebuild:
    project-name: test-lambuild
```

path | type | required | description
--- | --- | --- | ----
.repositories | []repository | true |

type: repository

path | type | required | default/example | description
--- | --- | --- | --- | ---
.name | string | true | `suzuki-shunsuke/test-lambuild` | repository full name `<repo_owner>/<repo_name>`
.hooks | []hook | true | |

If an event doesn't match any hook's condition, the event is ignored.

type: hook

path | type | required | default/example | description
--- | --- | --- | --- | ---
.config | string | false | `lambuild.yaml` | relative path from repository's root directory to the buildspec template file on the source repository
.if | string (expr's expression) | false | | if an event doesn't match the condition, the event is ignored. If this field is empty, no event is ignored

## Optional Environment Variables

name | default | example | description
--- | --- | --- | ---
LOG_LEVEL | `info` | `debug` | log level of `lambuild`. [logrus](https://github.com/sirupsen/logrus) is being used
BUILD_STATUS_CONTEXT | not specified | `AWS Codebuild ({{.event.Headers.Event}})` | [`build-status-config-override`'s context](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/codebuild/start-build.html)
ERROR_NOTIFICATION_TEMPLATE | | Error notification template

BUILD_STATUS_CONTEXT and ERROR_NOTIFICATION_TEMPLATE are rendered with Go's [text/template](https://golang.org/pkg/text/template/). [sprig functions](http://masterminds.github.io/sprig/) can be used.

BUILD_STATUS_CONTEXT's parameter

path | type | example | description
--- | --- | --- | ---
.event | Event | |
.pr | PullRequest | | Associated pull request
.repo | Repository | | Associated repository
.sha | string | | Associated commit SHA
.ref | string | `refs/heads/master` |

ERROR_NOTIFICATION_TEMPLATE's parameter

path | type | example | description
--- | --- | --- | ---
.Error | error | |
