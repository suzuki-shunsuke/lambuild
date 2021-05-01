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

func (handler *Handler) filter(filt bspec.LambuildFilter, webhook wh.Webhook, event wh.Event) (bool, error) {
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
	if f, err := handler.filterByMatchfile(filt.Ref, []string{event.Ref}); err != nil || !f {
		return false, err
	}
	if event.PRNum != 0 {
		if f, err := handler.filterByMatchfile(filt.File, event.ChangedFileNames); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.Label, event.Labels); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.Author, []string{event.PRAuthor}); err != nil || !f {
			return false, err
		}
		if f, err := handler.filterByMatchfile(filt.BaseRef, []string{event.BaseRef}); err != nil || !f {
			return false, err
		}
	}
	return true, nil
}
