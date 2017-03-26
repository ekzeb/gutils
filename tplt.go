package util

import (
	"html/template"
	"fmt"
)

func ParseTemplate(files ...string) *template.Template {
	return template.Must(
		template.New(files[0]).ParseFiles(files...),
	)
}

func ParseTemplate2(files []string, funcMap map[string]interface{}) *template.Template {
	return template.Must(
		template.New(files[0]).Funcs(funcMap).ParseFiles(files...),
	)
}
