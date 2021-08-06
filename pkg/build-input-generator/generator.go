package generator

import (
	"fmt"

	"github.com/sirupsen/logrus"
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
	"github.com/suzuki-shunsuke/lambuild/pkg/config"
	"github.com/suzuki-shunsuke/lambuild/pkg/domain"
	"github.com/suzuki-shunsuke/lambuild/pkg/template"
)

func GenerateInput(logE *logrus.Entry, buildStatusContext template.Template, data *domain.Data, buildspec bspec.Buildspec, repo config.Repository) (domain.BuildInput, error) {
	buildInput := domain.BuildInput{}

	if !buildspec.Lambuild.If.Empty() {
		f, err := buildspec.Lambuild.If.Run(data.Convert())
		if err != nil {
			return buildInput, fmt.Errorf("evaluate buildspec.Lambuild.If: %w", err)
		}
		if !f {
			return domain.BuildInput{
				Empty: true,
			}, nil
		}
	}

	return handleBuild(data, buildspec)
}
