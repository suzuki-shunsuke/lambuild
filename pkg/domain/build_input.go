package domain

import "github.com/aws/aws-sdk-go/service/codebuild"

type BuildInput struct {
	Build      *codebuild.StartBuildInput
	BatchBuild *codebuild.StartBuildBatchInput
	Batched    bool
	Empty      bool
}
