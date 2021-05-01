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

func (handler *Handler) generateListBuildspec(base bspec.Buildspec, elems []bspec.ListElement) bspec.Buildspec {
	arr := make([]bspec.ListElement, len(elems))
	for i, elem := range elems {
		elem.Lambuild = bspec.Lambuild{}
		arr[i] = elem
	}
	base.Batch.BuildList = arr
	return base
}
