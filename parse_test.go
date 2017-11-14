package httpio

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/schema"
)

type testInput struct {
	Name  string
	Image string `schema:"form-image"`
}

type testErr interface {
	Code() ErrType
}

func TestDefaultParse(t *testing.T) {
	for _, c := range []struct {
		Name     string
		HIO      *IO
		Method   string
		Path     string
		Body     string
		Headers  http.Header
		ExpInput *testInput
		ExpErr   ErrType
	}{
		{
			Name:     "plain GET should not decode as it has no content",
			HIO:      &IO{},
			Method:   http.MethodGet,
			ExpInput: &testInput{Name: ""},
		},
		{
			Name:     "GET with query should not decode as no form decoder is configured",
			HIO:      &IO{},
			Method:   http.MethodGet,
			Path:     "?name=foo",
			ExpInput: &testInput{Name: ""},
		},
		{
			Name:     "GET with query should decode with form decoder and custom field name",
			HIO:      &IO{FormDecoder: schema.NewDecoder()},
			Method:   http.MethodGet,
			Path:     "?name=foo&form-image=bar",
			ExpInput: &testInput{Name: "foo", Image: "bar"},
		},
		// {
		// 	Name:   "GET with query should decode should fail on invalid field",
		// 	HIO:    &IO{FormDecoder: schema.NewDecoder()},
		// 	Method: http.MethodGet,
		// 	Path:   "?name=foo&bogus=bogus&form-image=bar2",
		// 	ExpErr: ErrDecodeForm,
		// },
	} {
		t.Run(c.Name, func(t *testing.T) {
			var b io.Reader
			if c.Body != "" {
				b = bytes.NewBufferString(c.Body)
			}

			r, err := http.NewRequest(c.Method, c.Path, b)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			in := &testInput{}
			err = c.HIO.parse(r, in)
			if c.ExpErr != "" {
				eerr, ok := err.(testErr)
				if !ok {
					t.Fatal("parse error should be castable to test error")
				}

				if eerr.Code() != c.ExpErr {
					t.Fatalf("error code should be '%s' but got '%s'", c.ExpErr, eerr.Code())
				}
			}

			// if !reflect.DeepEqual(err, c.ExpErr) {
			//   if c.ExpErr.Code() ==
			//
			// 	t.Fatalf("Expected error: %#v, got: %#v", c.ExpErr, err)
			// }

			if !reflect.DeepEqual(in, c.ExpInput) {
				t.Fatalf("Expected input: %#v, got: %#v", c.ExpInput, in)
			}
		})
	}
}
