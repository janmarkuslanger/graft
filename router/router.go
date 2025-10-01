package router

import (
	"fmt"
	"net/http"
	"strings"
)

func New() *Router {
	mux := http.NewServeMux()
	return &Router{Mux: mux}
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

type HandlerFunc func(ctx Context)

type Router struct {
	Mux *http.ServeMux
}

func (r *Router) AddHandler(pattern string, handler HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler(Context{
			Writer:  w,
			Request: req,
		})
	})

	finalHandler := Chain(baseHandler, middlewares...)
	r.Mux.Handle(pattern, finalHandler)
}

func (r *Router) Static(urlPrefix, dir string, middlewares ...func(http.Handler) http.Handler) {
	// remove every / at the end so we dont have multiple /
	prefix := strings.TrimRight(urlPrefix, "/") + "/"
	h := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	final := Chain(h, middlewares...)
	r.Mux.Handle("GET "+prefix, final)
	r.Mux.Handle("HEAD "+prefix, final)
}

func (r *Router) Handle(pattern string, handler HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	fmt.Println("Serving route(s) with pattern: " + pattern)

	r.AddHandler(pattern, handler, middlewares...)

	if pattern[len(pattern)-1:] != "/" {
		r.AddHandler(pattern+"/", handler, middlewares...)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}
