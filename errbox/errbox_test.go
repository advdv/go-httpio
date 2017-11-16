package errbox

import (
	"errors"
	"testing"
)

type capitalizeInput struct {
	Name string `json:"name"`
}

type capitalizeOutput struct {
	Err  `json:"error,omitempty"`
	Name string `json:"name,omitempty"`
}

func TestErrorEmbedding(t *testing.T) {
	out := &capitalizeOutput{}
	if out.UnboxErr() != nil {
		t.Fatal("embedded error should now be nil")
	}

	out.BoxError(errors.New("some error"))
	if out.UnboxErr().Error() != "some error" {
		t.Fatal("embedded error should now not be nil")
	}
}

func TestBox(t *testing.T) {
	a := &capitalizeInput{}
	b := &capitalizeOutput{}

	v1, ok := Box(a, nil)
	if v1 != a || !ok {
		t.Fatal("expected value to be returned and OK")
	}

	err1 := errors.New("some error")
	v2, ok := Box(a, err1)
	if v2 != err1 || ok {
		t.Fatal("expected value to be returned and OK")
	}

	v3, ok := Box(b, err1)
	if v3 != b && !ok {
		t.Fatalf("expected value to be returned and OK: %v", v3)
	}
}
