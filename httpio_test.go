package httpio_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	for _, c := range []struct {
		Name      string
		Method    string
		Path      string
		Body      io.Reader
		Ctrl      *httpio.Ctrl
		Impl      func(context.Context, *testInput) (*testOutput, error)
		ExpBody   string
		ExpStatus int
		ExpHdr    http.Header
	}{
		{
			Name:      "GET an empty output struct",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ctrl:      httpio.NewCtrl(&val{validator.New()}, &encoding.JSON{}),
			ExpBody:   `{}` + "\n",
			ExpStatus: 200,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return &testOutput{}, nil
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(c.Method, c.Path, c.Body)
			func(w http.ResponseWriter, r *http.Request) {
				in := &testInput{}
				if render, valid := c.Ctrl.Handle(w, r, in, &testOutput{}); valid {
					render(c.Impl(r.Context(), in))
				}
			}(w, r)

			if w.Body.String() != c.ExpBody {
				t.Fatalf("expected resp body '%s', got: %s", c.ExpBody, w.Body.String())
			}

			if w.Code != c.ExpStatus {
				t.Fatalf("expected status '%d', got: %d", c.ExpStatus, w.Code)
			}

			if !reflect.DeepEqual(w.Header(), c.ExpHdr) {
				t.Fatalf("expected resp hdr '%v', got: %v", c.ExpHdr, w.Header())
			}

		})
	}

}
