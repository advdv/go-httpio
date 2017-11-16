package httpio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/advanderveer/go-httpio/encoding"
	validator "gopkg.in/go-playground/validator.v9"
)

type testInput struct {
	Name string
}

type testOutput struct{}

func testImpl(ctx context.Context, in *testInput) (*testOutput, error) {
	return &testOutput{}, nil
}

type val struct {
	v *validator.Validate
}

func (val *val) Validate(v interface{}) error {
	return val.v.Struct(v)
}

func TestBasicUsage(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	ctrl := httpio.NewCtrl(&encoding.JSON{})
	ctrl.V = &val{validator.New()}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := &testInput{}
		if render, valid := ctrl.Handle(w, r, in, &testOutput{}); valid {
			render(testImpl(r.Context(), in))
		}
	}

	h(w, r)

	if w.Body.String() != `{}`+"\n" {
		t.Fatal("got body:", w.Body.String())
	}

	//@TODO assert w

}
