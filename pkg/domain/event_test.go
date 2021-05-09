package domain_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
)

func TestData_CommitMessage(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		data  domain.Data
		exp   string
	}{
		{
			title: "normal",
			data: domain.Data{
				HeadCommitMessage: "hello",
			},
			exp: "hello",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			msg := d.data.CommitMessage()
			if msg != d.exp {
				t.Fatalf("got %s, wanted %s", msg, d.exp)
			}
		})
	}
}
