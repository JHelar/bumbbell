package mux

import (
	"net/http"
	"regexp"
)

const HttpMethodAny string = "*"

type HttpMux struct {
	prefix string
	routes map[string][]*Route
}

type Route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

func NewHttpMux(prefix string) *HttpMux {
	return &HttpMux{
		prefix: prefix,
		routes: make(map[string][]*Route),
	}
}

func (mux *HttpMux) Handle(pattern string, handler http.Handler) {
	mux.addRouteHandler(HttpMethodAny, pattern, handler)
}

func (mux *HttpMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) Get(pattern string, handler http.Handler) {
	mux.addRouteHandler(http.MethodGet, pattern, handler)
}

func (mux *HttpMux) GetFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Get(pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) Post(pattern string, handler http.Handler) {
	mux.addRouteHandler(http.MethodPost, pattern, handler)
}

func (mux *HttpMux) PostFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Post(pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) Put(pattern string, handler http.Handler) {
	mux.addRouteHandler(http.MethodPut, pattern, handler)
}

func (mux *HttpMux) PutFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Put(pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) Delete(pattern string, handler http.Handler) {
	mux.addRouteHandler(http.MethodDelete, pattern, handler)
}

func (mux *HttpMux) DeleteFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Delete(pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	routes := append(mux.routes[r.Method], mux.routes[HttpMethodAny]...)
	for _, route := range routes {
		if route.pattern.MatchString(r.URL.Path) {
			insertUrlParams(*route.pattern, r)

			route.handler.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

func (mux *HttpMux) Use(pattern string, middlewares ...func(http.Handler) http.Handler) *HttpMux {
	useMux := NewHttpMux(pattern)
	handler := http.Handler(useMux)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	mux.addRouteHandler(HttpMethodAny, pattern, handler)

	return useMux
}

func (mux *HttpMux) addRouteHandler(verb string, pattern string, handler http.Handler) {
	mux.routes[verb] = append(mux.routes[verb], &Route{
		pattern: regexp.MustCompile(mux.prefix + pattern),
		handler: handler,
	})
}

func insertUrlParams(pattern regexp.Regexp, r *http.Request) {
	match := pattern.FindStringSubmatch(r.URL.Path)

	// Hack to populate r.Form if encoding is set to multipart/form-data
	r.FormValue("")
	r.ParseForm()

	for i, name := range pattern.SubexpNames() {
		if i > 0 && i <= len(match) {
			r.Form.Add(name, match[i])
		}
	}
}
