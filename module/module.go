package module

import "github.com/janmarkuslanger/graft/router"

type Route[T any] struct {
	Path    string
	Method  string
	Handler func(ctx router.Context, deps T)
}

type Lifecycle[T any] struct {
	OnUse   func(deps *T)
	OnStart func(deps *T)
}

type Module[T any] struct {
	Name        string
	BasePath    string
	Deps        T
	Routes      []Route[T]
	Middlewares []router.Middleware
	Lifecycle   Lifecycle[T]
}

func (m *Module[T]) BuildRoutes(r router.Router) {
	deps := &m.Deps
	for i := range m.Routes {
		route := m.Routes[i]
		path := route.Method + " " + m.BasePath + route.Path
		handlerFunc := func(ctx router.Context) {
			route.Handler(ctx, *deps)
		}
		r.AddHandler(path, handlerFunc, m.Middlewares...)
	}
}

func (m *Module[T]) OnUse() {
	if m.Lifecycle.OnUse != nil {
		m.Lifecycle.OnUse(&m.Deps)
	}
}

func (m *Module[T]) OnStart() {
	if m.Lifecycle.OnStart != nil {
		m.Lifecycle.OnStart(&m.Deps)
	}
}
