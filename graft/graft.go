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
	Name() string
	BuildRoutes(router router.Router)
}

type Graft struct {
	router  *router.Router
	modules []Module
}

func (g *Graft) Run() {
	http.ListenAndServe(":8080", g.router)
}

func (g *Graft) UseModule(m Module) {
	g.modules = append(g.modules, m)
	m.BuildRoutes(*g.router)
}
