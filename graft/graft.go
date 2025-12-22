package graft

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/janmarkuslanger/graft/router"
)

func New() *Graft {
	r := router.New()

	return &Graft{
		router:   r,
		services: NewServices(),
	}
}

type Module interface {
	BuildRoutes(router router.Router)
}

type HookModule interface {
	Module
	OnUse()
	OnStart()
}

type Graft struct {
	router   *router.Router
	modules  []Module
	services *Services
}

var listenAndServe = http.ListenAndServe

func (g *Graft) Run() {
	g.startModules()
	listenAndServe(":8080", g.router)
}

func (g *Graft) UseModule(m Module) {
	if svcAware, ok := m.(ServiceAwareModule); ok {
		svcAware.SetServices(g.services)
	}
	if hooks, ok := m.(HookModule); ok {
		hooks.OnUse()
	}
	g.modules = append(g.modules, m)
	m.BuildRoutes(*g.router)
}

func (g *Graft) UseMiddleware(middlewares ...router.Middleware) {
	g.router.Use(middlewares...)
}

func (g *Graft) startModules() {
	for _, mod := range g.modules {
		if hooks, ok := mod.(HookModule); ok {
			hooks.OnStart()
		}
	}
}

func (g *Graft) Services() *Services {
	return g.services
}

func (g *Graft) RegisterService(name string, service any) {
	RegisterService(g.services, name, service)
}

func (g *Graft) GetService(name string) (any, bool) {
	return GetService[any](g.services, name)
}

func (g *Graft) MustGetService(name string) any {
	return MustGetService[any](g.services, name)
}

func (g *Graft) HasService(name string) bool {
	return HasService(g.services, name)
}

// Typed helpers without exposing the service bag.
func ServiceAs[T any](g *Graft, name string) (T, bool) {
	return GetService[T](g.services, name)
}

func MustServiceAs[T any](g *Graft, name string) T {
	return MustGetService[T](g.services, name)
}

type Services struct {
	mu     sync.RWMutex
	values map[string]any
}

func NewServices() *Services {
	return &Services{
		values: make(map[string]any),
	}
}

func RegisterService(s *Services, name string, service any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[name] = service
}

func GetService[T any](s *Services, name string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.values[name]
	if !ok {
		var zero T
		return zero, false
	}

	casted, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return casted, true
}

func MustGetService[T any](s *Services, name string) T {
	val, ok := GetService[T](s, name)
	if !ok {
		panic(fmt.Sprintf("service named %q not found", name))
	}
	return val
}

func HasService(s *Services, name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.values[name]
	return ok
}

type ServiceAwareModule interface {
	Module
	SetServices(*Services)
}
