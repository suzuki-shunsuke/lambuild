# lambuild configuration

`lambuild` gets configuration files from a source repository and generates buildspecs.
We call these files as `lambuild.yaml`, but we can rename freely.
The format of lambuild configuration file is based on buildspec.

* [Build specification reference for CodeBuild](https://docs.aws.amazon.com/codebuild/latest/userguide/build-spec-ref.html)
* [Batch build buildspec reference](https://docs.aws.amazon.com/codebuild/latest/userguide/batch-build-buildspec.html)

## For Batch Build

Please see [lambuild.yaml for Batch Build](lambuild-batch-yaml.md)

## Example

```yaml
---
version: 0.2
env:
  git-credential-helper: yes
lambuild:
  build-status-context: "foo"
  image: aws/codebuild/standard:1.0
  environment-type: LINUX_CONTAINER
  compute-type: BUILD_GENERAL1_SMALL
  privileged-mode: true
  debug-session: true
  git-clone-depth: 10
  report-build-status: true
  items:
    - FOO: foo
  env:
    variables:
      FOO: "{{.item.Foo}}"
phases:
  build:
    commands:
      - echo "foo"
```

## Reference

path | type | example | description
--- | --- | --- | ---
.lambuild.env.variables | `map[string](string expression)` | | build's environment variables. The environment variables of `.lambuild.env.variables` are passed by the override option
.lambuild.image | string | `alpine:3.13.0` |
.lambuild.git-clone-depth | int | `0` |
.lambuild.compute-type | string |  |
.lambuild.environment-type | string |  |
.lambuild.debug-session | bool | |
.lambuild.privileged-mode | bool | |
.lambuild.report-build-status | bool | |
.lambuild.items | []Item | |

* `type: bool expression` is a string whose evaluated result is a boolean
* `type: string expression` is a string whose evaluated result is a string
* `type: ExprList` is a list whose element is either `string` or `ExprElem`

## type: ExprElem

path | type | example | description
--- | --- | --- | ---
.value | string | |
.if | bool expression | |

## type: Item

path | type | example | description
--- | --- | --- | ---
.if | bool expression | |
.env | `map[string](string expression)` | | build's environment variables
.build-status-context | template string | `"foo ({{.event.Headers.Event}})"` |
.image | string | `aws/codebuild/standard:5.0` |
.compute-type | string | `BUILD_GENERAL1_SMALL` |
.environment-type | string | `LINUX_CONTAINER` |
.param | `map[string]interface{}` | | a parameter `item` of template and expression
