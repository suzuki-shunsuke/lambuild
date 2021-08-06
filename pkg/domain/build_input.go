package domain

import "github.com/aws/aws-sdk-go/service/codebuild"

type BuildInput struct {
	Builds []*codebuild.StartBuildInput
	Empty  bool
}
