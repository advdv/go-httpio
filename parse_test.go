package httpio

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/schema"
	"gopkg.in/go-playground/validator.v9"
)

type testInput struct {
	Name  string
	Image string `schema:"form-image" json:"json-image" validate:"min=10"`
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
		ExpInput *testInput //we expect the body to be parsed into the input like this
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
		{
			Name:     "GET with invalid query should give an error",
			HIO:      &IO{FormDecoder: schema.NewDecoder()},
			Method:   http.MethodGet,
			Path:     "?name=fo%uo&form-image=bar",
			ExpInput: &testInput{Name: "", Image: ""},
			ExpErr:   ErrParseForm,
		},
		{
			Name:     "GET with query should decode should fail on invalid field",
			HIO:      &IO{FormDecoder: schema.NewDecoder()},
			Method:   http.MethodGet,
			Path:     "?name=foo&bogus=bogus&form-image=bar2",
			ExpInput: &testInput{Name: "foo", Image: "bar2"},
			ExpErr:   ErrDecodeForm,
		},
		{
			Name:     "POST with json but no decoders",
			HIO:      &IO{FormDecoder: schema.NewDecoder()},
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Body:     `{"json-image": "bar3"}`,
			ExpInput: &testInput{Name: "foo", Image: "bar2"},
			ExpErr:   ErrNoDecoders,
		},
		{
			Name:     "POST with json but with decoder should give unsupported content",
			HIO:      NewJSON(schema.NewDecoder(), nil),
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Body:     `{"json-image": "bar3"}`,
			ExpInput: &testInput{Name: "foo", Image: "bar2"},
			ExpErr:   ErrUnsupportedContent,
		},
		{
			Name:     "POST with json should overwrite image but not name",
			HIO:      NewJSON(schema.NewDecoder(), nil),
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Headers:  http.Header{"Content-Type": []string{"application/json"}},
			Body:     `{"json-image": "bar3"}`,
			ExpInput: &testInput{Name: "foo", Image: "bar3"},
			ExpErr:   "",
		},
		{
			Name:     "POST with json with invalid json",
			HIO:      NewJSON(schema.NewDecoder(), nil),
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Headers:  http.Header{"Content-Type": []string{"application/json"}},
			Body:     `{"json-image": "bar3}`,
			ExpInput: &testInput{Name: "foo", Image: "bar2"},
			ExpErr:   ErrDecodeBody,
		},
		{
			Name:     "POST with json with validation",
			HIO:      NewJSON(schema.NewDecoder(), validator.New()),
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Headers:  http.Header{"Content-Type": []string{"application/json"}},
			Body:     `{"json-image": "bar3"}`,
			ExpInput: &testInput{Name: "foo", Image: "bar3"},
			ExpErr:   ErrInputValidation,
		},
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

			r.Header = c.Headers

			in := &testInput{}
			err = c.HIO.parse(r, in)
			if !IsErrType(err, c.ExpErr) {
				t.Fatalf("expected error %v to be code '%v'", err, c.ExpErr)
			}

			if !reflect.DeepEqual(in, c.ExpInput) {
				t.Fatalf("Expected input: %#v, got: %#v", c.ExpInput, in)
			}
		})
	}
}
