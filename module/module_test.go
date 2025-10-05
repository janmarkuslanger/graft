package module_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmarkuslanger/graft/module"
	"github.com/janmarkuslanger/graft/router"
)

func TestModule_New(t *testing.T) {
	name := "User"
	m := module.New[any](name, "/user", struct{}{})
	if m.Name() != name {
		t.Fatalf("expected module name to be %q", name)
	}
}

func TestModule_AddRoute(t *testing.T) {
	m := module.New[any]("User", "/user", struct{}{})
	m.AddRoute(module.Route[any]{
		Path:   "/login",
		Method: "GET",
		Handler: func(ctx router.Context, deps any) {
			ctx.Writer.WriteHeader(http.StatusOK)
			ctx.Writer.Write([]byte("Hello World"))
		},
	})

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

	m := module.New("User", "/user", Deps{UserService: UserService{}})
	m.AddRoute(module.Route[Deps]{
		Path:   "/login",
		Method: "GET",
		Handler: func(ctx router.Context, deps Deps) {
			ctx.Writer.WriteHeader(http.StatusOK)
			ctx.Writer.Write([]byte(deps.UserService.Login()))
		},
	})

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
