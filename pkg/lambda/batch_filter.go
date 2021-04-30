package lambda

import (
	"strings"

	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	wh "github.com/suzuki-shunsuke/lambuild/pkg/webhook"
)

func (handler *Handler) filterByMatchfile(rawCondition string, arr []string) (bool, error) {
	if rawCondition == "" {
		return true, nil
	}
	cond, err := handler.MatchfileParser.ParseConditions(strings.Split(strings.TrimSpace(rawCondition), "\n"))
	if err != nil {
		return false, err
	}
	f, err := handler.MatchfileParser.Match(arr, cond)
	if err != nil {
		return false, err
	}
	if !f {
		return false, nil
	}
	return true, nil
}

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
	if f, err := handler.filterByMatchfile(filt.Ref, []string{body.Ref}); err != nil || !f {
		return false, err
	}
	if pr.Number != 0 {
		if f, err := handler.filterByMatchfile(filt.File, pr.FileNames); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.Label, pr.Labels); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.Author, []string{pr.PullRequest.GetUser().GetLogin()}); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.BaseRef, []string{pr.PullRequest.Base.GetRef()}); err != nil || !f {
			return false, err
		}
	}
	return true, nil
}
