package domain_test

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v35/github"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
)

func TestData_GetCommit(t *testing.T) {
	t.Parallel()
	d := domain.NewData()
	d.Commit.Set(&github.Commit{
		Message: github.String("hello"),
	})
	data := []struct {
		title string
		data  domain.Data
		exp   *github.Commit
	}{
		{
			title: "normal",
			data:  d,
			exp: &github.Commit{
				Message: github.String("hello"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			commit := d.data.GetCommit()
			if !reflect.DeepEqual(*commit, *d.exp) {
				t.Fatalf("got %+v, wanted %+v", *commit, *d.exp)
			}
		})
	}
}

func TestData_GetPRNumber(t *testing.T) {
	t.Parallel()
	domainData := domain.NewData()
	domainData.PullRequest.Number.Set(5)
	data := []struct {
		title string
		data  domain.Data
		exp   int
	}{
		{
			title: "normal",
			data:  domainData,
			exp:   5,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			prNumber := d.data.GetPRNumber()
			if d.exp != prNumber {
				t.Fatalf("got %d, wanted %d", prNumber, d.exp)
			}
		})
	}
}

func TestData_GetPR(t *testing.T) {
	t.Parallel()
	domainData := domain.NewData()
	domainData.PullRequest.PullRequest.Set(&github.PullRequest{
		Number: github.Int(5),
	})
	data := []struct {
		title string
		data  domain.Data
		exp   *github.PullRequest
	}{
		{
			title: "normal",
			data:  domainData,
			exp: &github.PullRequest{
				Number: github.Int(5),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pr := d.data.GetPR()
			if !reflect.DeepEqual(d.exp, pr) {
				t.Fatalf("got %+v, wanted %+v", pr, d.exp)
			}
		})
	}
}

func TestData_GetPRFiles(t *testing.T) {
	t.Parallel()
	domainData := domain.NewData()
	domainData.PullRequest.Files.Set([]*github.CommitFile{
		{
			Filename: github.String("foo"),
		},
	})
	data := []struct {
		title string
		data  domain.Data
		exp   []*github.CommitFile
	}{
		{
			title: "normal",
			data:  domainData,
			exp: []*github.CommitFile{
				{
					Filename: github.String("foo"),
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			files := d.data.GetPRFiles()
			if !reflect.DeepEqual(d.exp, files) {
				t.Fatalf("got %+v, wanted %+v", files, d.exp)
			}
		})
	}
}
