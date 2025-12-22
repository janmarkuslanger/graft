package router_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestRouter_Static_ServesFiles(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "hello.txt")
	contents := []byte("hello static")
	if err := os.WriteFile(filePath, contents, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := router.New()
	r.Static("/assets/", dir)

	getReq := httptest.NewRequest(http.MethodGet, "/assets/hello.txt", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getRec.Code)
	}
	if getRec.Body.String() != string(contents) {
		t.Fatalf("expected body %q, got %q", contents, getRec.Body.String())
	}

	headReq := httptest.NewRequest(http.MethodHead, "/assets/hello.txt", nil)
	headRec := httptest.NewRecorder()
	r.ServeHTTP(headRec, headReq)

	if headRec.Code != http.StatusOK {
		t.Fatalf("expected HEAD status 200, got %d", headRec.Code)
	}
	if headRec.Body.Len() != 0 {
		t.Fatalf("expected empty HEAD body, got %q", headRec.Body.String())
	}
}

func TestRouter_Use_GlobalMiddlewares(t *testing.T) {
	r := router.New()

	var calls []string

	r.Use(func(ctx router.Context, next router.HandlerFunc) {
		calls = append(calls, "global-before")
		next(ctx)
		calls = append(calls, "global-after")
	})

	r.AddHandler("GET /test", func(ctx router.Context) {
		calls = append(calls, "handler")
		ctx.Writer.WriteHeader(http.StatusNoContent)
	}, func(ctx router.Context, next router.HandlerFunc) {
		calls = append(calls, "route-before")
		next(ctx)
		calls = append(calls, "route-after")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}

	expected := []string{
		"global-before",
		"route-before",
		"handler",
		"route-after",
		"global-after",
	}

	if len(calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(calls))
	}

	for i := range calls {
		if calls[i] != expected[i] {
			t.Errorf("at index %d expected %q, got %q", i, expected[i], calls[i])
		}
	}
}
