package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestChain(t *testing.T) {
	order := make([]string, 0, 3)

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "base")
	})

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1")
			next.ServeHTTP(w, r)
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2")
			next.ServeHTTP(w, r)
		})
	}

	chained := Chain(base, m1, m2)
	chained.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com", nil))

	expected := []string{"m1", "m2", "base"}
	for i, want := range expected {
		if order[i] != want {
			t.Fatalf("expected order[%d] to be %q, got %q", i, want, order[i])
		}
	}
}

func TestRouter_AddHandler(t *testing.T) {
	r := New()

	var ctx Context
	callCount := 0

	r.AddHandler("GET /hello", func(got Context) {
		ctx = got
		callCount++
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/hello", nil)
	req.Host = "GET "
	req.URL = &url.URL{Path: "/hello"}
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if callCount != 1 {
		t.Fatalf("expected handler to be called once, got %d", callCount)
	}
	if ctx.Request != req {
		t.Fatalf("expected context to contain original request")
	}
	if ctx.Writer == nil {
		t.Fatalf("expected context writer to be set")
	}
}

func TestRouter_Static(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "index.txt")
	if err := os.WriteFile(file, []byte("static-content"), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := New()

	middlewareCalls := 0
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			middlewareCalls++
			next.ServeHTTP(w, req)
		})
	}

	r.Static("/static/", dir, middleware)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/static/index.txt", nil)
	req.Host = "GET "
	req.URL = &url.URL{Path: "/static/index.txt"}
	res := httptest.NewRecorder()

	handler, _ := r.Mux.Handler(req)
	handler.ServeHTTP(res, req)

	if res.Body.String() != "static-content" {
		t.Fatalf("expected static file content, got %q", res.Body.String())
	}
	if middlewareCalls != 1 {
		t.Fatalf("expected middleware to run once, got %d", middlewareCalls)
	}
}

func TestRouter_Handle(t *testing.T) {
	r := New()

	callCount := 0
	r.Handle("GET /api", func(ctx Context) {
		callCount++
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api", nil)
	req.Host = "GET "
	req.URL = &url.URL{Path: "/api"}
	r.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/api/", nil)
	req2.Host = "GET "
	req2.URL = &url.URL{Path: "/api/"}
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount != 2 {
		t.Fatalf("expected handler to be called twice, got %d", callCount)
	}
}
