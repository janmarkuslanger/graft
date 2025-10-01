package router_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	routerpkg "github.com/janmarkuslanger/graft/router"
)

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func head(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodHead, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestNew(t *testing.T) {
	r := routerpkg.New()
	if r == nil || r.Mux == nil {
		t.Fatalf("New returned nil")
	}
}

func TestAddHandler_GET(t *testing.T) {
	r := routerpkg.New()
	r.AddHandler("GET /hello", func(ctx routerpkg.Context) {
		_, _ = io.WriteString(ctx.Writer, "ok")
	})

	rec := get(t, r, "/hello")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestAddHandler_Middleware(t *testing.T) {
	r := routerpkg.New()
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-MW", "1")
			next.ServeHTTP(w, req)
		})
	}
	r.AddHandler("GET /mw", func(ctx routerpkg.Context) {
		_, _ = io.WriteString(ctx.Writer, "mw")
	}, mw)

	rec := get(t, r, "/mw")
	if rec.Header().Get("X-MW") != "1" {
		t.Fatalf("missing middleware header")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandle_TrailingSlash(t *testing.T) {
	r := routerpkg.New()
	count := 0
	h := func(ctx routerpkg.Context) {
		count++
		_, _ = io.WriteString(ctx.Writer, "hit")
	}
	r.Handle("GET /foo", h)

	rec1 := get(t, r, "/foo")
	if rec1.Code != http.StatusOK {
		t.Fatalf("status=%d", rec1.Code)
	}
	rec2 := get(t, r, "/foo/")
	if rec2.Code != http.StatusOK {
		t.Fatalf("status=%d", rec2.Code)
	}
	if count != 2 {
		t.Fatalf("hits=%d", count)
	}
}

func TestStatic_GET_HEAD(t *testing.T) {
	dir := t.TempDir()
	want := "hello\n"
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}

	r := routerpkg.New()
	r.Static("/static", dir)

	recG := get(t, r, "/static/f.txt")
	if recG.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", recG.Code, recG.Body.String())
	}
	if recG.Body.String() != want {
		t.Fatalf("body=%q", recG.Body.String())
	}

	recH := head(t, r, "/static/f.txt")
	if recH.Code != http.StatusOK {
		t.Fatalf("status=%d", recH.Code)
	}
	if recH.Body.Len() != 0 {
		t.Fatalf("head has body bytes=%d", recH.Body.Len())
	}
}

func TestServeHTTP(t *testing.T) {
	r := routerpkg.New()
	r.AddHandler("GET /ping", func(ctx routerpkg.Context) {
		_, _ = io.WriteString(ctx.Writer, "pong")
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "pong" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}
