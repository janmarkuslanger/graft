package graft_test

import (
	"testing"

	"github.com/janmarkuslanger/graft/graft"
)

type TestModule struct{}

func (m TestModule) Name() string {
	return "Test"
}

func TestGraft_New(t *testing.T) {
	app := graft.New()
	if app == nil {
		t.Errorf("Expect an app instance but got %v", app)
	}
}
