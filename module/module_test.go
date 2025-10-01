package module_test

import (
	"testing"

	"github.com/janmarkuslanger/graft/module"
)

func TestModule_New(t *testing.T) {
	name := "User"
	m := module.New[any](name, "/user", struct{}{})
	if m.Name() != name {
		t.Fatalf("Expected module name to be: %v", name)
	}
}
