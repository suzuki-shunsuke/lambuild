package lambda

import (
	bspec "github.com/suzuki-shunsuke/lambuild/pkg/buildspec"
)

func (handler *Handler) generateBuildspec(base bspec.Buildspec, graphElems []bspec.GraphElement) bspec.Buildspec {
	arr := make([]bspec.GraphElement, len(graphElems))
	for i, elem := range graphElems {
		elem.Lambuild = bspec.Lambuild{}
		arr[i] = elem
	}
	base.Batch.BuildGraph = arr
	return base
}
