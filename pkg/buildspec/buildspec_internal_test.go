package buildspec

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildspec_filter(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		buildspec Buildspec
		param     interface{}
		exp       map[string]interface{}
	}{
		{
			title: "minimum",
			exp: map[string]interface{}{
				"batch": Batch{},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			m, err := d.buildspec.filter(d.param)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
