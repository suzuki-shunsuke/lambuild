# lambuild.yaml

`lambuild` gets a configuration file from a source repository and generates buildspec.
We call this configuration file as `lambuild.yaml`, but we can change the filename and path freely.

`lambuild` generates buildspec based on `lambuild.yaml`, so the format of `lambuild.yaml` is based on buildspec.

* [Build specification reference for CodeBuild](https://docs.aws.amazon.com/codebuild/latest/userguide/build-spec-ref.html)
* [Batch build buildspec reference](https://docs.aws.amazon.com/codebuild/latest/userguide/batch-build-buildspec.html)

## Example

build-list

```yaml
---
version: 0.2
env:
  git-credential-helper: yes
lambuild:
  env:
    variables:
      LAMBUILD_WEBHOOK_BODY: "event.Body"
      LAMBUILD_WEBHOOK_EVENT: "event.Headers.Event"
      LAMBUILD_WEBHOOK_DELIVERY: "event.Headers.Delivery"
      LAMBUILD_HEAD_COMMIT_MSG: "getCommitMessage()"
batch:
  build-list:
    - identifier: foo
      buildspec: foo/buildspec.yaml
      if: 'any(getPRFileNames(), {# startsWith "foo/"})'
    - identifier: bar
      buildspec: buildspec/renovate.yaml
      if: 'any(getPRFileNames(), {# == "renovate.json"})'
```

build-graph

```yaml
---
version: 0.2
batch:
  build-graph:
    - identifier: build
      buildspec: foo/build.yaml
      if: 'any(getPRFileNames(), {# startsWith "foo/"})'
    - identifier: deploy
      buildspec: foo/deploy.yaml
      if: 'any(getPRFileNames(), {# startsWith "foo/"})'
      depend-on:
        - build
```

build-matrix

```yaml
---
version: 0.2
batch:
  build-matrix:
    static:
      ignore-failure: false
      env:
        type: LINUX_CONTAINER
        image: aws/codebuild/standard:5.0
        privileged-mode: true
    dynamic:
      buildspec:
        - buildspec/foo.yaml
        - value: buildspec/pull_request.yaml
          if: event.Headers.Event == "pull_request"
      env:
        compute-type:
          - BUILD_GENERAL1_SMALL
          - value: BUILD_GENERAL1_MEDIUM
            if: ref == "refs/heads/master"
        image:
          - aws/codebuild/amazonlinux2-x86_64-standard:3.0
          - value: aws/codebuild/windows-base:2019-1.0
            if: ref == "refs/heads/master"
        variables:
          FOO:
            - hello
            - value: pull_request
              if: event.Headers.Event == "pull_request"
            - value: push
              if: event.Headers.Event == "push"
```

## Reference

path | type | example | description
--- | --- | --- | ---
.lambuild.env.variables | `map[string]string`. The value of map is expr's expression | | build's environment variables
.batch.build-list[].if | string (expr's expression) | |
.batch.build-graph[].if | string (expr's expression) | |
.batch.build-matrix.dynamic.buildspec | ExprList | |
.batch.build-matrix.dynamic.env.compute-type | ExprList | |
.batch.build-matrix.dynamic.env.image | ExprList | |
.batch.build-matrix.dynamic.env.variables | `map[string]ExprList` | |

`type: ExprList` is a list whose element is either `string` or `ExprElem`.

type: ExprElem

path | type | example | description
--- | --- | --- | ---
.value | string | |
.if | string (expr's expression) | |

---

When `build-list` and `build-graph`'s all builds are removed by `if` condition, then no build is started.

In case of `build-matrix`, when all configuration (buildspec, compute-type, image, variables) are removed by `if` condition, then no build is started.

After filtering builds by `if` condition, if the number of remaining builds is one,
then `lambuild` calls CodeBuild `StartBuild` API instead of `StartBuildBatch` API,
which means `lambuild` starts not Batch Build but Build.
Batch Build has some overhead, so `lambuild` starts Build instead of Batch Build to decrease the Build time.
