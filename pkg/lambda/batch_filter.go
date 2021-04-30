package lambda

import (
	"strings"

	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
)

func (handler *Handler) filter(filt bspec.LambuildFilter, webhook wh.Webhook, body wh.Body, pr PullRequest) (bool, error) {
	if len(filt.Event) != 0 {
		matched := false
		for _, ev := range filt.Event {
			if ev == webhook.Headers.Event {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}
	if pr.Number != 0 {
		if filt.File != "" {
			cond, err := handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(filt.File), "\n"))
			if err != nil {
				return false, err
			}
			f, err := handler.MatchfileParser.Match(pr.FileNames, cond)
			if err != nil {
				return false, err
			}
			if !f {
				return false, nil
			}
		}
		if filt.Label != "" {
			cond, err := handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(filt.Label), "\n"))
			if err != nil {
				return false, err
			}
			f, err := handler.MatchfileParser.Match(pr.Labels, cond)
			if err != nil {
				return false, err
			}
			if !f {
				return false, nil
			}
		}
	}
	return true, nil
}
