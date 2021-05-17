# lambuild.yaml for Batch Build

## Note

Batch Build supports to specify `buildspec` as the following.

```yaml
  build-list:
    - identifier: foo
      buildspec: foo/buildspec.yaml
```

But `lambuild` doesn't check this file, so in this file `lambuild` specific configuration can't be used. 

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
.batch.build-list[].if | string expression | |
.batch.build-graph[].if | string expression | |
.batch.build-matrix.dynamic.buildspec | ExprList | |
.batch.build-matrix.dynamic.env.compute-type | ExprList | |
.batch.build-matrix.dynamic.env.image | ExprList | |
.batch.build-matrix.dynamic.env.variables | `map[string]ExprList` | |

## Specification to generate buildspec

When Batch Build's all builds are removed by `if` condition, then no build is started.

After filtering builds by `if` condition, if the number of remaining builds is one,
then `lambuild` calls CodeBuild `StartBuild` API instead of `StartBuildBatch` API,
which means `lambuild` starts not Batch Build but Build.
Batch Build has some overhead, so `lambuild` starts Build instead of Batch Build to decrease the Build time.
In that case, each build's properties are passed by the override option.

In case of `build-graph`, there are dependencies between builds.
After filtering builds by `if` condition, if some builds which the build `A` depends on are removed then the build `A` also is removed.

For example, in case of the following configuration the build `deploy` isn't run because `deploy` depends on `build` and `build` isn't run.

```yaml
  build-graph:
    - identifier: build
      if: "false"
    - identifier: deploy
      depend-on:
        - build
```
