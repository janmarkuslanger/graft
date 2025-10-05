package module_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmarkuslanger/graft/module"
	"github.com/janmarkuslanger/graft/router"
)

func TestModule_AddRoute(t *testing.T) {
	m := module.Module[any]{
		Name:     "User",
		BasePath: "/user",
		Routes: []module.Route[any]{
			module.Route[any]{
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

	m := module.Module[Deps]{
		Name:     "User",
		BasePath: "/user",
		Routes: []module.Route[Deps]{
			module.Route[Deps]{
				Path:   "/login",
				Method: "GET",
				Handler: func(ctx router.Context, deps Deps) {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte(deps.UserService.Login()))
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
	if rr.Body.String() != "Logged in" {
		t.Errorf("expected body 'Logged in', got %q", rr.Body.String())
	}
}
