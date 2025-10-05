package module

import "github.com/janmarkuslanger/graft/router"

func New[T any](name string, basePath string, deps T) *Module[T] {
	return &Module[T]{name: name, basePath: basePath, deps: deps}
}

type Route[T any] struct {
	Path    string
	Method  string
	Handler func(ctx router.Context, deps T)
}

type Module[T any] struct {
	name     string
	basePath string
	deps     T
	routes   []Route[T]
}

func (m *Module[T]) Name() string {
	return m.name
}

func (m *Module[T]) AddRoute(route Route[T]) {
	m.routes = append(m.routes, route)

}

func (m *Module[T]) BuildRoutes(r router.Router) {
	for _, route := range m.routes {
		deps := m.deps
		path := route.Method + " " + m.basePath + route.Path
		handlerFunc := func(ctx router.Context) {
			route.Handler(ctx, deps)
		}
		r.AddHandler(path, handlerFunc)
	}
}
