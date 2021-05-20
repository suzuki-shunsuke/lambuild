package buildspec_test

import (
	"testing"

	"github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"gopkg.in/yaml.v2"
)

func TestExprList_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   string
	}{
		{
			title: "normal",
			src: `- foo
- value: yoo
  if: "true"
`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			list := buildspec.ExprList{}
			if err := yaml.Unmarshal([]byte(d.src), &list); err != nil {
				t.Fatal(err)
			}
			if len(list) == 0 {
				t.Fatal("list is empty")
			}
		})
	}
}

func TestMatrix_Empty(t *testing.T) {
	t.Parallel()
	data := []struct {
		title  string
		matrix buildspec.Matrix
		exp    bool
	}{
		{
			title: "minimum",
			exp:   true,
		},
		{
			title: "buildspec",
			matrix: buildspec.Matrix{
				Dynamic: buildspec.MatrixDynamic{
					Buildspec: buildspec.ExprList{"foo.yaml"},
				},
			},
			exp: false,
		},
		{
			title: "compute-type",
			matrix: buildspec.Matrix{
				Dynamic: buildspec.MatrixDynamic{
					Env: buildspec.MatrixDynamicEnv{
						ComputeType: buildspec.ExprList{"BUILD_GENERAL1_SMALL"},
					},
				},
			},
			exp: false,
		},
		{
			title: "image",
			matrix: buildspec.Matrix{
				Dynamic: buildspec.MatrixDynamic{
					Env: buildspec.MatrixDynamicEnv{
						Image: buildspec.ExprList{"alpine"},
					},
				},
			},
			exp: false,
		},
		{
			title: "envvars",
			matrix: buildspec.Matrix{
				Dynamic: buildspec.MatrixDynamic{
					Env: buildspec.MatrixDynamicEnv{
						Variables: map[string]buildspec.ExprList{
							"FOO": {"foo"},
						},
					},
				},
			},
			exp: false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if d.exp != d.matrix.Empty() {
				t.Fatalf(`wanted %v, got %v`, d.exp, d.matrix.Empty())
			}
		})
	}
}
