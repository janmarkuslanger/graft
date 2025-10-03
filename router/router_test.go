package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmarkuslanger/graft/router"
)

func TestRouter_AddHandler_Success(t *testing.T) {
	r := router.New()
	r.AddHandler("POST /test", func(ctx router.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("Hello World"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got %q", rr.Body.String())
	}
}

func TestRouter_AddHandler_NotFound(t *testing.T) {
	r := router.New()
	r.AddHandler("/test", func(ctx router.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("Hello World"))
	})

	req := httptest.NewRequest("GET", "/not-test", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestRouter_AddHandler_WrongMethod(t *testing.T) {
	r := router.New()
	r.AddHandler("POST /test", func(ctx router.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("Hello World"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestRouter_Handle_AddWithSlash(t *testing.T) {
	r := router.New()
	r.Handle("GET /test", func(ctx router.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("Hello World"))
	})

	routes := []string{
		"/test",
		"/test/",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	}
}
