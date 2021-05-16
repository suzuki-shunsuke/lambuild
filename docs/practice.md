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
