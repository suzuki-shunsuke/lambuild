package lambda

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v35/github"
)

func Test_extractLabelNames(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		input []*github.Label
		exp   []string
	}{
		{
			title: "normal",
			input: []*github.Label{
				{
					Name: github.String("foo"),
				},
			},
			exp: []string{"foo"},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			names := extractLabelNames(d.input)
			if !reflect.DeepEqual(names, d.exp) {
				t.Fatalf("got %+v, wanted %+v", names, d.exp)
			}
		})
	}
}
