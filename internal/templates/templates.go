package templates

import (
	"dumbbell/internal/environment"
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
	"isDev": func() bool {
		return environment.GetEnvironment() == environment.Development
	},
}

var Page = template.Must(template.Must(template.New("pageTemplates").Funcs(templateFunctions).ParseGlob("templates/pages/*.html")).ParseGlob("templates/partials/*.html"))
var Htmx = template.Must(template.Must(template.New("htmxTemplates").Funcs(templateFunctions).ParseGlob("templates/htmx/*.html")).ParseGlob("templates/partials/*.html"))
var Partials = template.Must(template.New("partials").Funcs(templateFunctions).ParseGlob("templates/partials/*.html"))
var StartWorkout = template.Must(Partials.New("startWorkout").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	{{ template "exercise" .Exercise }}
`))
var PickWorkout = template.Must(Partials.New("pickWorkout").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	<div hx-swap-oob="afterbegin:body">{{ template "header" .Header }}</div>
	{{ template "workoutContainer" . }}
`))
var NextExercise = template.Must(Partials.New("nextExercise").Parse(`
	{{ template "exerciseSets" . }}
`))
var Home = template.Must(Partials.New("home").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	<div hx-swap-oob="afterbegin:body">{{ template "header" .Header }}</div>
	{{ template "dashboardContainer" . }}
`))
var Login = template.Must(Partials.New("login").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	{{ template "loginContainer" . }}
`))
var Signup = template.Must(Partials.New("signup").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	{{ template "signupContainer" . }}
`))
var Settings = template.Must(Partials.New("settings").Parse(`
	<title>{{ .Title }}</title>
	<div hx-swap-oob="delete:#page-header"></div>
	<div hx-swap-oob="afterbegin:body">{{ template "header" .Header }}</div>
	{{ template "settingsContainer" . }}
`))
var AlertBanner = template.Must(Partials.New("userCredentialsError").Parse(`
	{{ template "alertBanner" . }}
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
