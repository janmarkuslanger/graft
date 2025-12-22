package graft

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/janmarkuslanger/graft/module"
	"github.com/janmarkuslanger/graft/router"
)

type testModule struct {
	built bool
}

type hookModule struct {
	used    int
	started int
}

func (m *testModule) Name() string {
	return "Test"
}

func (m *testModule) BuildRoutes(r router.Router) {
	m.built = r.Mux != nil
}

func (m *hookModule) BuildRoutes(r router.Router) {}

func (m *hookModule) OnUse() {
	m.used++
}

func (m *hookModule) OnStart() {
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

func TestGraft_Hooks(t *testing.T) {
	app := New()
	mod := &hookModule{}

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

func TestGraft_GlobalAndModuleMiddleware(t *testing.T) {
	app := New()

	app.UseMiddleware(func(ctx router.Context, next router.HandlerFunc) {
		ctx.Writer.Header().Add("X-Global", "true")
		next(ctx)
	})

	mod := &module.Module[struct{}]{
		BasePath: "/ext",
		Middlewares: []router.Middleware{
			func(ctx router.Context, next router.HandlerFunc) {
				ctx.Writer.Header().Add("X-Module", "true")
				next(ctx)
			},
		},
		Routes: []module.Route[struct{}]{
			{
				Method: http.MethodGet,
				Path:   "/",
				Handler: func(ctx router.Context, _ struct{}) {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte("ok"))
				},
			},
		},
	}

	app.UseModule(mod)

	req := httptest.NewRequest(http.MethodGet, "/ext/", nil)
	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 from module, got %d", rr.Code)
	}
	if rr.Header().Get("X-Global") != "true" {
		t.Fatalf("expected global middleware to run")
	}
	if rr.Header().Get("X-Module") != "true" {
		t.Fatalf("expected module middleware to run")
	}
}

func TestGraft_ServiceHelpers(t *testing.T) {
	app := New()

	type DB struct {
		URL string
	}

	db := DB{URL: "postgres://localhost:5432/db"}

	app.RegisterService("db", db)

	if !app.HasService("db") {
		t.Fatalf("expected service to be registered")
	}

	got, ok := ServiceAs[DB](app, "db")
	if !ok {
		t.Fatalf("expected to get service")
	}
	if got.URL != db.URL {
		t.Fatalf("expected URL %q, got %q", db.URL, got.URL)
	}

	must := MustServiceAs[DB](app, "db")
	if must.URL != db.URL {
		t.Fatalf("expected MustGetService to return registered value")
	}
}

type serviceAwareModule struct {
	services *Services
}

func (m *serviceAwareModule) BuildRoutes(r router.Router) {
	// no-op for test
}

func (m *serviceAwareModule) SetServices(s *Services) {
	m.services = s
}

func TestGraft_ServiceAwareModule(t *testing.T) {
	app := New()
	app.RegisterService("db", "db-conn")

	mod := &serviceAwareModule{}
	app.UseModule(mod)

	if mod.services == nil {
		t.Fatalf("expected services to be injected")
	}

	val, ok := GetService[string](mod.services, "db")
	if !ok || val != "db-conn" {
		t.Fatalf("expected module to see registered service, got ok=%v val=%q", ok, val)
	}
}
