package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmarkuslanger/graft/router"
)

func TestMiddleware_Chain(t *testing.T) {
	var callOrder []string

	mw1 := func(ctx router.Context, next router.HandlerFunc) {
		callOrder = append(callOrder, "mw1-before")
		next(ctx)
		callOrder = append(callOrder, "mw1-after")
	}

	mw2 := func(ctx router.Context, next router.HandlerFunc) {
		callOrder = append(callOrder, "mw2-before")
		next(ctx)
		callOrder = append(callOrder, "mw2-after")
	}

	handler := func(ctx router.Context) {
		callOrder = append(callOrder, "handler")
		ctx.Writer.WriteHeader(http.StatusNoContent)
	}

	chained := router.Chain(handler, mw1, mw2)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	chained(router.Context{Writer: w, Request: req})

	expectedOrder := []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}

	if len(callOrder) != len(expectedOrder) {
		t.Fatalf("expected %d calls, got %d", len(expectedOrder), len(callOrder))
	}

	for i := range callOrder {
		if callOrder[i] != expectedOrder[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expectedOrder[i], callOrder[i])
		}
	}

	if w.Result().StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Result().StatusCode)
	}
}
