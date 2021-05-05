package template

import (
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func Compile(s string) (*template.Template, error) {
	tpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse a template: %w", err)
	}
	return tpl, nil
}
