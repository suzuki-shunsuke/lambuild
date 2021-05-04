# Expression

`lambuild` uses the Expression Engine [antonmedv/expr](https://github.com/antonmedv/expr) to filter events and generate the buildspec dynamically.

The following parameters are passed to expressions.

.path | type | example | description
--- | --- | --- | ---
.event | Event | |
.repo | Repository | |
.sha | string | |
.ref | string | |
.getCommit | `func() *github.Commit` | |
.getCommitMessage | `func() string` | |
.getPR | `func() *github.PullRequest` | | get an associated pull request
.getPRNumber | `func() int` | | get an associated pull request number
.getPRFiles | `func() []*github.CommitFile` | | get associated pull request files
.getPRFileNames | `func() []string` | | get associated pull request file paths
.getPRLabelNames | `func() []string` | | get associated pull request label names
.regexp.match | `func(pattern, s string) bool` | | returns true if `s` matches the regular expression `pattern`. Go's [regexp#MatchString](https://golang.org/pkg/regexp/#MatchString) is used

Please see [go-github's document](https://pkg.go.dev/github.com/google/go-github/v35/github) too.

* [*github.Commit](https://pkg.go.dev/github.com/google/go-github/v35/github#Commit)
* [*github.PullRequest](https://pkg.go.dev/github.com/google/go-github/v35/github#PullRequest)
* [*github.CommitFile](https://pkg.go.dev/github.com/google/go-github/v35/github#CommitFile)

_Note that the above go-github version may be different from actual version which lambuild uses, because it is difficult to maintain the document properly. Please check [go.mod](../go.mod)._

To avoid unneeded GitHub API call, `lambuild` doesn't call API until the function is called at the expression, and the result is cached in the Lambda Function's request scope.

This is the reason why the type of parameters like `getPRFileNames` is function.
