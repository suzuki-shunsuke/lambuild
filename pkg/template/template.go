package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Template struct {
	template *template.Template
}

func (tpl *Template) Empty() bool {
	return tpl.template == nil
}

func (tpl *Template) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var a string
	if err := unmarshal(&a); err != nil {
		return err
	}
	t, err := compile(a)
	if err != nil {
		return err
	}
	tpl.template = t
	return nil
}

func (tpl *Template) Execute(param interface{}) (string, error) {
	if tpl.template == nil {
		return "", nil
	}
	buf := &bytes.Buffer{}
	if err := tpl.template.Execute(buf, param); err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return buf.String(), nil
}

func New(s string) (Template, error) {
	tpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(s)
	if err != nil {
		return Template{}, fmt.Errorf("parse a template: %w", err)
	}
	return Template{template: tpl}, nil
}

func compile(s string) (*template.Template, error) {
	tpl, err := template.New("_").Funcs(sprig.TxtFuncMap()).Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse a template: %w", err)
	}
	return tpl, nil
}
