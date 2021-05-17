# Practice

## GitHub API Rate exceeded

If you encounter GitHub API Rate exceeded problem, please check the following solution.

* Use dedicated GitHub User for Personal Access Token
  * https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting
* Filter requests properly by lambuild hook configuration before getting `lambuild.yaml` from source repositories
* Use either push or pull_request event at the same time
* reduce the number of API call to get `lambuild.yaml` from source repositories
  * If the number of `lambuild.yaml` is one, specify not the parent directory but the file path
  * Reduce the number of `lambuild.yaml`

## Conflict of GitHub Status Context

If you run multiple builds per commit, you have to prevent the conflict of the GitHub Status Context.

https://docs.github.com/en/rest/reference/repos#statuses

lambuild supports to specify the build status context.
To prevent the conflict, it is good to include the following information in the context.

* Event type like `push` and `pull_request`
* AWS Account ID
* Region
* CodeBuild Project Name

e.g.

```yaml
  build-status-context: "{{.item.name}} ({{.event.Headers.Event}}) {{.aws.Codebuild.ProjectName}} {{.aws.Region}} {{.aws.AccountID}}"
```

But unfortunately, AWS CodeBuild's `StartBuildBatch` API doesn't support to specify the build status context unlike `StartBuild` API.

* [StartBuild API](https://docs.aws.amazon.com/codebuild/latest/APIReference/API_StartBuild.html#API_StartBuild_RequestSyntax)
* [StartBuildBatch API](https://docs.aws.amazon.com/codebuild/latest/APIReference/API_StartBuildBatch.html)
