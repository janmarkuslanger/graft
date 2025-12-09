package graft

import (
	"net/http"

	"github.com/janmarkuslanger/graft/router"
)

func New() *Graft {
	r := router.New()

	return &Graft{
		router: r,
	}
}

type Module interface {
	BuildRoutes(router router.Router)
}

type LifecycleModule interface {
	Module
	OnUse()
	OnStart()
}

type Graft struct {
	router  *router.Router
	modules []Module
}

var listenAndServe = http.ListenAndServe

func (g *Graft) Run() {
	g.startModules()
	listenAndServe(":8080", g.router)
}

func (g *Graft) UseModule(m Module) {
	if lifecycle, ok := m.(LifecycleModule); ok {
		lifecycle.OnUse()
	}
	g.modules = append(g.modules, m)
	m.BuildRoutes(*g.router)
}

func (g *Graft) startModules() {
	for _, mod := range g.modules {
		if lifecycle, ok := mod.(LifecycleModule); ok {
			lifecycle.OnStart()
		}
	}
}
