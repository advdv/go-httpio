package handling_test

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
	"github.com/gorilla/schema"
)

type testInput struct {
	Name  string
	Image string `schema:"form-image" json:"json-image" validate:"min=10"`
}

func TestDefaultParse(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Handling *handling.H
		Method   string
		Path     string
		Body     string
		Headers  http.Header
		ExpInput *testInput //we expect the body to be parsed into the input like this
		ExpErr   error
	}{
		{
			Name: "plain GET should not decode as it has no content",
			Handling: handling.NewH(
				encoding.NewStack(&encoding.JSON{}),
			),
			Method:   http.MethodGet,
			ExpInput: &testInput{Name: ""},
		},
		{
			Name: "GET with query should not decode as no form decoder is configured",
			Handling: handling.NewH(
				encoding.NewStack(&encoding.JSON{}),
			),
			Method:   http.MethodGet,
			Path:     "?name=foo",
			ExpInput: &testInput{Name: ""},
		},
		{
			Name: "GET with query should decode with form decoder and custom field name",
			Handling: handling.NewH(
				encoding.NewStack(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			),
			Method:   http.MethodGet,
			Path:     "?name=foo&form-image=bar",
			ExpInput: &testInput{Name: "foo", Image: "bar"},
		},
		// {
		// 	Name:     "GET with invalid query should give an error",
		// 	Method:   http.MethodGet,
		// 	Path:     "?name=fo%uo&form-image=bar",
		// 	ExpInput: &testInput{Name: "", Image: ""},
		// },
		// {
		// 	Name:     "GET with query should decode should fail on invalid field",
		// 	Method:   http.MethodGet,
		// 	Path:     "?name=foo&bogus=bogus&form-image=bar2",
		// 	ExpInput: &testInput{Name: "foo", Image: "bar2"},
		// },
		// {
		// 	Name: "POST with json but no decoders",
		// 	Handling: handling.NewH(
		// 		encoding.NewStack(encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
		// 	),
		// 	Method:   http.MethodPost,
		// 	Path:     "?name=foo&form-image=bar2",
		// 	Body:     `{"json-image": "bar3"}`,
		// 	ExpInput: &testInput{Name: "foo", Image: "bar2"},
		// },
		// {
		// 	Name:     "POST with json but with decoder should give unsupported content",
		// 	Method:   http.MethodPost,
		// 	Path:     "?name=foo&form-image=bar2",
		// 	Body:     `{"json-image": "bar3"}`,
		// 	ExpInput: &testInput{Name: "foo", Image: "bar2"},
		// },
		{
			Name: "POST with json should overwrite image but not name",
			Handling: handling.NewH(
				encoding.NewStack(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			),
			Method:   http.MethodPost,
			Path:     "?name=foo&form-image=bar2",
			Headers:  http.Header{"Content-Type": []string{"application/json"}},
			Body:     `{"json-image": "bar3"}`,
			ExpInput: &testInput{Name: "foo", Image: "bar3"},
		},
		// {
		// 	Name:     "POST with json with invalid json",
		// 	Method:   http.MethodPost,
		// 	Path:     "?name=foo&form-image=bar2",
		// 	Headers:  http.Header{"Content-Type": []string{"application/json"}},
		// 	Body:     `{"json-image": "bar3}`,
		// 	ExpInput: &testInput{Name: "foo", Image: "bar2"},
		// },
		// {
		// 	Name:     "POST with json with validation",
		// 	Method:   http.MethodPost,
		// 	Path:     "?name=foo&form-image=bar2",
		// 	Headers:  http.Header{"Content-Type": []string{"application/json"}},
		// 	Body:     `{"json-image": "bar3"}`,
		// 	ExpInput: &testInput{Name: "foo", Image: "bar3"},
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

			r.Header = c.Headers
			in := &testInput{}
			err = c.Handling.Parse(r, in)
			if !reflect.DeepEqual(err, c.ExpErr) {
				t.Fatalf("Expected error: %#v, got: %#v", c.ExpErr, err)
			}

			if !reflect.DeepEqual(in, c.ExpInput) {
				t.Fatalf("Expected input: %#v, got: %#v", c.ExpInput, in)
			}
		})
	}
}
