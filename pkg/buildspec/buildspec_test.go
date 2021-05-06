package buildspec_test

import (
	"reflect"
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func TestBuildspec_MarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		buildspec buildspec.Buildspec
		exp       map[string]interface{}
	}{
		{
			title: "normal",
			buildspec: buildspec.Buildspec{
				Batch: buildspec.Batch{
					BuildGraph: []buildspec.GraphElement{
						{
							Identifier: "build",
							Buildspec:  "build.yaml",
						},
					},
				},
				Map: map[string]interface{}{
					"version": "0.2",
				},
			},
			exp: map[string]interface{}{
				"version": "0.2",
				"batch": map[interface{}]interface{}{
					"build-graph": []interface{}{
						map[interface{}]interface{}{
							"identifier": "build",
							"buildspec":  "build.yaml",
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			b, err := yaml.Marshal(&d.buildspec)
			if err != nil {
				t.Fatal(err)
			}
			m := map[string]interface{}{}
			if err := yaml.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.exp, m) {
				t.Fatalf("got %+v, wanted %+v", m, d.exp)
			}
		})
	}
}
