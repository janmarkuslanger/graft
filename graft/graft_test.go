package graft

import (
	"net/http"
	"sync"
	"testing"

	"github.com/janmarkuslanger/graft/router"
)

type testModule struct {
	built bool
}

type lifecycleModule struct {
	used    int
	started int
}

func (m *testModule) Name() string {
	return "Test"
}

func (m *testModule) BuildRoutes(r router.Router) {
	m.built = r.Mux != nil
}

func (m *lifecycleModule) BuildRoutes(r router.Router) {}

func (m *lifecycleModule) OnUse() {
	m.used++
}

func (m *lifecycleModule) OnStart() {
	m.started++
}

func TestGraft_New(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatalf("expected an app instance")
	}
}

func TestGraft_UseModule(t *testing.T) {
	app := New()
	mod := &testModule{}

	app.UseModule(mod)

	if !mod.built {
		t.Fatalf("expected BuildRoutes to be called")
	}
}

func TestGraft_Run(t *testing.T) {
	app := New()
	app.UseModule(&testModule{})

	var (
		mu          sync.Mutex
		addr        string
		seenHandler *router.Router
	)

	original := listenAndServe
	listenAndServe = func(a string, h http.Handler) error {
		mu.Lock()
		defer mu.Unlock()
		addr = a
		if r, ok := h.(*router.Router); ok {
			seenHandler = r
		}
		return nil
	}
	t.Cleanup(func() { listenAndServe = original })

	app.Run()

	mu.Lock()
	defer mu.Unlock()
	if addr != ":8080" {
		t.Fatalf("expected addr :8080, got %q", addr)
	}
	if seenHandler == nil {
		t.Fatalf("expected listenAndServe to receive the router")
	}
}

func TestGraft_LifecycleHooks(t *testing.T) {
	app := New()
	mod := &lifecycleModule{}

	app.UseModule(mod)

	if mod.used != 1 {
		t.Fatalf("expected OnUse to run exactly once, got %d", mod.used)
	}

	original := listenAndServe
	listenAndServe = func(addr string, h http.Handler) error {
		if mod.started != 1 {
			t.Fatalf("expected OnStart before server startup, got %d", mod.started)
		}
		return nil
	}
	t.Cleanup(func() { listenAndServe = original })

	app.Run()
}
