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

func (m *testModule) Name() string {
	return "Test"
}

func (m *testModule) BuildRoutes(r router.Router) {
	m.built = r.Mux != nil
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
