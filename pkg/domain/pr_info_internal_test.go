package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v36/github"
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
			if diff := cmp.Diff(names, d.exp); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func Test_extractPRFileNames(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		input []*github.CommitFile
		exp   map[string]struct{}
	}{
		{
			title: "normal",
			input: []*github.CommitFile{
				{
					Filename: github.String("foo"),
				},
				{
					Filename:         github.String("zoo"),
					PreviousFilename: github.String("bar"),
				},
			},
			exp: map[string]struct{}{
				"foo": {},
				"zoo": {},
				"bar": {},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			names := extractPRFileNames(d.input)
			if len(d.exp) != len(names) {
				t.Fatalf("got %d, wanted %d", len(names), len(d.exp))
			}
			for _, name := range names {
				if _, ok := d.exp[name]; !ok {
					t.Fatalf("%s isn't included", name)
				}
			}
		})
	}
}
