package module_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmarkuslanger/graft/module"
	"github.com/janmarkuslanger/graft/router"
)

func TestModule_WithRoute(t *testing.T) {
	m := module.Module[any]{
		Name:     "User",
		BasePath: "/user",
		Routes: []module.Route[any]{
			{
				Path:   "/login",
				Method: "GET",
				Handler: func(ctx router.Context, deps any) {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte("Hello World"))
				},
			},
		},
	}

	r := router.New()
	m.BuildRoutes(*r)

	req := httptest.NewRequest("GET", "/user/login", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got %q", rr.Body.String())
	}
}

type UserService struct{}

func (s UserService) Login() string {
	return "Logged in"
}

func TestModule_WithDeps(t *testing.T) {
	type Deps struct {
		UserService UserService
	}

	var (
		onUseCalled   bool
		onStartCalled bool
	)

	m := module.Module[Deps]{
		Name:     "User",
		BasePath: "/user",
		Routes: []module.Route[Deps]{
			{
				Path:   "/login",
				Method: "GET",
				Handler: func(ctx router.Context, deps Deps) {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte(deps.UserService.Login()))
				},
			},
		},
		Hooks: module.Hooks[Deps]{
			OnUse: func(deps *Deps) {
				onUseCalled = true
				deps.UserService = UserService{}
			},
			OnStart: func(deps *Deps) {
				if deps.UserService.Login() == "" {
					t.Fatalf("expected hooks to see deps updates")
				}
				onStartCalled = true
			},
		},
	}

	r := router.New()
	m.OnUse()
	m.BuildRoutes(*r)
	m.OnStart()

	req := httptest.NewRequest("GET", "/user/login", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Logged in" {
		t.Errorf("expected body 'Logged in', got %q", rr.Body.String())
	}
	if !onUseCalled {
		t.Fatalf("expected OnUse hook to run")
	}
	if !onStartCalled {
		t.Fatalf("expected OnStart hook to run")
	}
}

func TestModule_WithMiddleware(t *testing.T) {
	m := module.Module[any]{
		Name:     "User",
		BasePath: "/user",
		Routes: []module.Route[any]{
			{
				Path:   "/login",
				Method: "GET",
				Handler: func(ctx router.Context, deps any) {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte("Hello World"))
				},
			},
		},
		Middlewares: []router.Middleware{
			func(ctx router.Context, next router.HandlerFunc) {
				ctx.Writer.WriteHeader(http.StatusOK)
				ctx.Writer.Write([]byte("I am a middleware"))
			},
		},
	}

	r := router.New()
	m.BuildRoutes(*r)

	req := httptest.NewRequest("GET", "/user/login", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "I am a middleware" {
		t.Errorf("expected body 'I am a middleware', got %q", rr.Body.String())
	}
}

func TestModule_HooksAllowDepsMutation(t *testing.T) {
	type Deps struct {
		Message string
	}

	m := module.Module[Deps]{
		BasePath: "/msg",
		Deps:     Deps{Message: "initial"},
		Routes: []module.Route[Deps]{{
			Path:   "/hello",
			Method: "GET",
			Handler: func(ctx router.Context, deps Deps) {
				ctx.Writer.Write([]byte(deps.Message))
			},
		}},
	}

	r := router.New()
	m.BuildRoutes(*r)

	// Directly mutate dependencies after routes are built to verify handler sees updated value.
	m.Deps.Message = "updated after hooks"

	req := httptest.NewRequest("GET", "/msg/hello", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Body.String() != "updated after hooks" {
		t.Fatalf("expected updated dependency to be visible in handler, got %q", rr.Body.String())
	}
}
