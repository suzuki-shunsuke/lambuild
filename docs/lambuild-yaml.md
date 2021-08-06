# lambuild configuration

`lambuild` gets configuration files from a source repository and generates buildspecs.
We call these files as `lambuild.yaml`, but we can rename freely.
The format of lambuild configuration file is based on buildspec.

* [Build specification reference for CodeBuild](https://docs.aws.amazon.com/codebuild/latest/userguide/build-spec-ref.html)

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
      - command: echo "main"
        if: |
          ref == "refs/heads/main" # main branch
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
.phases.install.commands | [][Command](#type-command) | |
.phases.pre_build.commands | [][Command](#type-command) | |
.phases.build.commands | [][Command](#type-command) | |
.phases.post_build.commands | [][Command](#type-command) | |

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

## type: Command

string or following struct

path | type | example | description
--- | --- | --- | ---
.command | string | `echo "hello"` |
.if | bool expression | |

Command is ignored when the evaluated result of `.if` is `false`.

e.g.

```yaml
phases:
  build:
    commands:
      - echo "run always"
      - command: bash release.sh
        if: |
          ref == "refs/heads/main" # main branch
```

## Run multiple builds with items

e.g.

```yaml
version: 0.2
env:
  git-credential-helper: yes
lambuild:
  build-status-context: "{{.item.name}} ({{.event.Headers.Event}})"
  items:
    - param:
        name: foo
    - param:
        name: bar
  env:
    variables:
      NAME: "{{.item.name}}"
phases:
  build:
    commands:
      - echo "NAME: $NAME"
```

When `.lambuild.items` is specified, a build is run per the element of `.lambuild.items`.
In case of the above example, two builds (`foo` and `bar`) are run.
And `param` field is passed to the expression and template as the variable `item`.

## Environment Variables

Please see [Custom Environment Variables](environment-variables.md).
