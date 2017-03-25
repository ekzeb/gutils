package util

import (
	"html/template"
)

func ParseTemplate(files []string, funcMap map[string]interface{}) *template.Template {
	return template.Must(
		template.New(files[0]).Funcs(funcMap).ParseFiles(files...),
	)
}
