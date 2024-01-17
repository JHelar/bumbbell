package templates

import (
	"html/template"
	"io"
	"reflect"
)

var templateFunctions = template.FuncMap{
	"isLast": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
	"isNotLast": func(x int, a interface{}) bool {
		return x != reflect.ValueOf(a).Len()-1
	},
}
var page = template.Must(template.Must(template.New("pageTemplates").Funcs(templateFunctions).ParseGlob("templates/pages/*.html")).ParseGlob("templates/partials/*.html"))
var htmx = template.Must(template.Must(template.New("htmxTemplates").Funcs(templateFunctions).ParseGlob("templates/htmx/*.html")).ParseGlob("templates/partials/*.html"))

func ExecutePageTemplate(wr io.Writer, templateName string, data any) error {
	return page.ExecuteTemplate(wr, templateName, data)
}

func ExecuteHtmxTemplate(wr io.Writer, templateName string, data any) error {
	return htmx.ExecuteTemplate(wr, templateName, data)
}
