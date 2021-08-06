# Practice

## GitHub API Rate exceeded

If you encounter GitHub API Rate exceeded problem, please check the following solution.

* Use dedicated GitHub User for Personal Access Token
  * https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting
* Filter requests properly by lambuild hook configuration before getting `lambuild.yaml` from source repositories
* Use either push or pull_request event at the same time
* reduce the number of API calls to get `lambuild.yaml` from source repositories
  * If the number of `lambuild.yaml` is one, instead of specifying the parent directory specify the file path
  * Reduce the number of `lambuild.yaml`

## Conflict of GitHub Status Context

If you run multiple builds per commit, you have to prevent the conflict of the GitHub Status Context.

https://docs.github.com/en/rest/reference/repos#statuses

lambuild supports specifying the build status context.
To prevent the conflict, it is good to include the following information in the context.

* Event type like `push` and `pull_request`
* AWS Account ID
* Region
* CodeBuild Project Name

e.g.

```yaml
  build-status-context: "{{.item.name}} ({{.event.Headers.Event}}) {{.aws.Codebuild.ProjectName}} {{.aws.Region}} {{.aws.AccountID}}"
```
