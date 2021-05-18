# Custom Environment Variables

[CodeBuild provides built-in environment variables](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html),
but it is inconvenient that when we retry builds some environment variables such as `CODEBUILD_WEBHOOK_BASE_REF` are lost.

lambuild supports to define environment variables with data of GitHub Webhook Event and associated pull request,
and these environment variables aren't lost when we retry builds.

e.g.

```yaml
lambuild:
  env:
    variables:
      LAMBUILD_WEBHOOK_BASE_REF: "{{.event.Payload.PullRequest.Base.Ref}}" # CODEBUILD_WEBHOOK_BASE_REF
      LAMBUILD_WEBHOOK_HEAD_REF: "{{.event.Payload.PullRequest.Head.Ref}}" # CODEBUILD_WEBHOOK_HEAD_REF
      LAMBUILD_WEBHOOK_EVENT: "{{.event.Headers.Event}}" # CODEBUILD_WEBHOOK_EVENT
```
