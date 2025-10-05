package module

import "github.com/janmarkuslanger/graft/router"

type Route[T any] struct {
	Path    string
	Method  string
	Handler func(ctx router.Context, deps T)
}

type Module[T any] struct {
	Name        string
	BasePath    string
	Deps        T
	Routes      []Route[T]
	Middlewares []router.Middleware
}

func (m *Module[T]) BuildRoutes(r router.Router) {
	for _, route := range m.Routes {
		deps := m.Deps
		path := route.Method + " " + m.BasePath + route.Path
		handlerFunc := func(ctx router.Context) {
			route.Handler(ctx, deps)
		}
		r.AddHandler(path, handlerFunc, m.Middlewares...)
	}
}
