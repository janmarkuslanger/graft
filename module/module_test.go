package module_test

import (
	"net/http/httptest"
	"net/url"
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

func TestModule_AddDependenciesAndBuildRoutes(t *testing.T) {
	type deps struct {
		value string
	}

	m := module.New[deps]("Test", "/test", deps{value: "old"})
	newDeps := &deps{value: "new"}
	m.AddDepencies(newDeps)

	var (
		handlerCalled bool
		received      deps
	)

	m.AddRoute(module.Route[deps]{
		Path:   "/info",
		Method: "GET",
		Handler: func(ctx router.Context, d deps) {
			handlerCalled = true
			received = d
		},
	})

	r := router.New()
	m.BuildRoutes(*r)

	req := httptest.NewRequest("GET", "http://example.com/test/info", nil)
	req.Host = "GET "
	req.URL = &url.URL{Path: "/test/info"}
	res := httptest.NewRecorder()

	handler, _ := r.Mux.Handler(req)
	handler.ServeHTTP(res, req)

	if !handlerCalled {
		t.Fatalf("expected route handler to be called")
	}
	if received.value != "new" {
		t.Fatalf("expected deps value 'new', got %q", received.value)
	}
}
