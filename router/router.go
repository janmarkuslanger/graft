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

func (r *Router) AddHandler(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	finalHandler := Chain(handler, middlewares...)
	r.Mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		finalHandler(Context{
			Writer:  w,
			Request: req,
		})
	})
}

func (r *Router) Static(urlPrefix, dir string, middlewares ...Middleware) {
	// remove every / at the end so we dont have multiple /
	prefix := strings.TrimRight(urlPrefix, "/") + "/"
	h := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	r.AddHandler("GET "+prefix, func(ctx Context) {
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}, middlewares...)
	r.AddHandler("HEAD "+prefix, func(ctx Context) {
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}, middlewares...)
}

func (r *Router) Handle(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fmt.Println("Serving route(s) with pattern: " + pattern)

	r.AddHandler(pattern, handler, middlewares...)

	if pattern[len(pattern)-1:] != "/" {
		r.AddHandler(pattern+"/", handler, middlewares...)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}
