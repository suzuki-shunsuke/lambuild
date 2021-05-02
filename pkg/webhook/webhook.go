package webhook

type Webhook struct {
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

type Headers struct {
	Event     string `json:"x-github-event"`
	Delivery  string `json:"x-github-delivery"`
	Signature string `json:"x-hub-signature-256"`
}

type Event struct {
	RepoFullName string
	RepoName     string

	RepoOwner string

	Ref              string
	PRAuthor         string
	BaseRef          string
	HeadRef          string
	ChangedFileNames []string
	Labels           []string

	PRNum int

	HeadCommitMessage string

	SHA string

	ChangedFiles int
}
