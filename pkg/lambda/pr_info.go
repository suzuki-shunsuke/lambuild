package lambda

import (
	"context"

	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
)

func (handler *Handler) getPRNum(ctx context.Context, webhook wh.Webhook, body wh.Body) (int, error) {
	if webhook.Headers.Event == "push" {
		return 0, nil
	}
	return 0, nil
}
