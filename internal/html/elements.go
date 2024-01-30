package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type ElementNode func(w io.Writer) error
type Element[T any] func(props T, children ...any) ElementNode
type ElementData struct {
	Attrs    template.HTMLAttr
	Children template.HTML
}

type ElementProps interface {
	String() string
}

func propsToAttrs(props any) string {
	attrs := []string{}

	switch reflect.TypeOf(props).Kind() {
	case reflect.Map:
		for key, value := range props.(map[string]string) {
			attrs = append(attrs, fmt.Sprintf("%s=\"%s\"", key, value))
		}
		break
	case reflect.Struct:
		log.Printf("Cannot struct prop type.")
		break
	default:
		log.Printf("Cannot handle prop type.")
		break
	}

	return strings.Join(attrs, " ")
}

func createElementTag[TProps interface{}](tag string) Element[TProps] {
	tmpl := template.Must(template.New(tag).Parse(fmt.Sprintf("<%s {{ .Attrs }}>{{ .Children }}</%s>", tag, tag)))

	return func(props TProps, children ...any) ElementNode {
		return func(w io.Writer) error {
			childWriter := bytes.NewBufferString("")
			for _, child := range children {
				if reflect.TypeOf(child).Kind() == reflect.Func {
					child.(ElementNode)(childWriter)
				} else if reflect.TypeOf(child).Kind() == reflect.String {
					childWriter.WriteString(child.(string))
				}
			}

			attrs := template.HTMLAttr(propsToAttrs(props))

			return tmpl.Execute(w, ElementData{
				Attrs:    attrs,
				Children: template.HTML(childWriter.String()),
			})
		}
	}
}

func RenderResponse(w http.ResponseWriter, node ElementNode) error {
	w.WriteHeader(http.StatusOK)
	return node(w)
}

var Div = createElementTag[map[string]string]("div")
var A = createElementTag[map[string]string]("a")
var Html = createElementTag[map[string]string]("html")
var Body = createElementTag[map[string]string]("body")
var Head = createElementTag[map[string]string]("head")

func Layout(children ...any) ElementNode {
	return Html(nil,
		Body(
			map[string]string{
				"class": "bg-zinc-800",
			},
			children...,
		),
	)
}
