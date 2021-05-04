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
.getPR | `func() *github.PullRequest` | |
.getPRNumber | `func() int` | |
.getPRFiles | `func() []*github.CommitFile` | |
.getPRFileNames | `func() []string` | |
.getPRLabelNames | `func() []string` | |
.regexp.match | `func(pattern, s string) bool` | |
