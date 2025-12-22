package graft

import "testing"

type svcType struct {
	Name string
}

func TestServices_ProvideAndGet(t *testing.T) {
	services := NewServices()
	expected := svcType{Name: "demo"}

	RegisterService(services, "svc", expected)

	got, ok := GetService[svcType](services, "svc")
	if !ok {
		t.Fatalf("expected to retrieve service")
	}
	if got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	if !HasService(services, "svc") {
		t.Fatalf("expected Has to report true")
	}
}

func TestServices_MustGet_PanicsOnMissing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic when service missing")
		}
	}()

	MustGetService[svcType](NewServices(), "missing")
}
