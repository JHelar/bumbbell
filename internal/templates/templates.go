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

var Page = template.Must(template.Must(template.New("pageTemplates").Funcs(templateFunctions).ParseGlob("templates/pages/*.html")).ParseGlob("templates/partials/*.html"))
var Htmx = template.Must(template.Must(template.New("htmxTemplates").Funcs(templateFunctions).ParseGlob("templates/htmx/*.html")).ParseGlob("templates/partials/*.html"))
var Partials = template.Must(template.New("partials").Funcs(templateFunctions).ParseGlob("templates/partials/*.html"))
var StartWorkout = template.Must(Partials.New("startWorkout").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	<div hx-swap-oob="outerHTML:#container">{{ template "exercise" .Exercise }}</div>
`))
var PickWorkout = template.Must(Partials.New("pickWorkout").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="afterbegin:body">{{ template "header" .Header }}</div>
	<div hx-swap-oob="outerHTML:#container">{{ template "workoutContainer" . }}</div>
`))
var NextExercise = template.Must(Partials.New("nextExercise").Parse(`
	{{ template "exerciseSets" . }}
`))
var Home = template.Must(Partials.New("home").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="afterbegin:body">{{ template "header" .Header }}</div>
	<div hx-swap-oob="outerHTML:#container">{{ template "dashboardContainer" . }}</div>
`))

func ExecutePageTemplate(wr io.Writer, templateName string, data any) error {
	return Page.ExecuteTemplate(wr, templateName, data)
}

func ExecuteHtmxTemplate(wr io.Writer, templateName string, data any) error {
	return Htmx.ExecuteTemplate(wr, templateName, data)
}

func New(name string) *template.Template {
	return template.New(name)
}
