package webhook

type Webhook struct {
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

type Body struct {
	Ref        string     `json:"ref"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Sender     Sender     `json:"sender"`
	HeadCommit HeadCommit `json:"head_commit"`
}

type Headers struct {
	Event     string `json:"x-github-event"`
	Delivery  string `json:"x-github-delivery"`
	Signature string `json:"x-hub-signature-256"`
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type Sender struct {
	Login string `json:"login"`
}

type Pusher struct {
	Name string `json:"name"`
}

type HeadCommit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
