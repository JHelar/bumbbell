package mux

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const HttpMethodAny string = "*"

type MuxHandler interface {
	MatchesPath(path string) bool
	InsertUrlParams(r *http.Request)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type HttpMux struct {
	prefix string
	routes map[string][]MuxHandler
}

type Route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

func (r *Route) MatchesPath(path string) bool {
	isMatch := r.pattern.MatchString(path)
	if isMatch {
		log.Printf("Matches Route Path: pattern=%s path=%s, isMatch=%t", r.pattern.String(), path, isMatch)
	}
	return isMatch
}

func (r *Route) InsertUrlParams(req *http.Request) {
	match := r.pattern.FindStringSubmatch(req.URL.Path)

	// Hack to populate r.Form if encoding is set to multipart/form-data
	req.FormValue("")
	req.ParseForm()

	for i, name := range r.pattern.SubexpNames() {
		if i > 0 && i <= len(match) {
			req.Form.Add(name, match[i])
		}
	}
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func NewHttpMux(prefix string) *HttpMux {
	return &HttpMux{
		prefix: prefix,
		routes: make(map[string][]MuxHandler),
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
		if route.MatchesPath(r.URL.Path) {
			route.InsertUrlParams(r)

			route.ServeHTTP(w, r)
			return
		}
	}

	log.Printf("No route matched: path=%s", r.URL.Path)

	http.NotFound(w, r)
}

func (mux *HttpMux) MatchesPath(path string) bool {
	isMatch := strings.HasPrefix(path, mux.prefix)

	if isMatch {
		log.Printf("Matches Mux Path: path=%s, isMatch=%t", path, isMatch)
	}

	return isMatch
}

func (mux *HttpMux) InsertUrlParams(r *http.Request) {
	// Stub
}

func (mux *HttpMux) Use(pattern string, middlewares ...func(http.Handler) http.Handler) *HttpMux {
	useMux := NewHttpMux(mux.prefix + pattern)
	handler := http.Handler(useMux)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	mux.addRoute(HttpMethodAny, &Route{
		pattern: regexp.MustCompile(fmt.Sprintf("^%s%s", mux.prefix, pattern)),
		handler: handler,
	})

	return useMux
}

func (mux *HttpMux) addRoute(verb string, handler MuxHandler) {
	mux.routes[verb] = append(mux.routes[verb], handler)
}

func (mux *HttpMux) addRouteHandler(verb string, pattern string, handler http.Handler) {
	mux.addRoute(verb, &Route{
		pattern: regexp.MustCompile(fmt.Sprintf("^%s%s$", mux.prefix, pattern)),
		handler: handler,
	})
}
