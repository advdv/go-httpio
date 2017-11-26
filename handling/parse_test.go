package handling_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
	"github.com/gorilla/schema"
	validate "gopkg.in/go-playground/validator.v9"
)

type val struct{ v *validate.Validate }

func (val val) Validate(v interface{}) error { return val.v.Struct(v) }

type testInput struct {
	Name  string
	Image string `schema:"form-image" json:"json-image" validate:"ascii"`
}

func (i *testInput) Validate() error {
	if i.Image == "invalid" {
		return errors.New("invalid image")
	}
	return nil
}

func TestDefaultParse(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Handling *handling.H
		// Validator handling.Validator
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
			if fmt.Sprint(c.ExpErr) != fmt.Sprint(err) {
				t.Fatalf("Expected error: %#v, got: %#v", fmt.Sprint(c.ExpErr), fmt.Sprint(err))
			}

			if !reflect.DeepEqual(in, c.ExpInput) {
				t.Fatalf("Expected input: %#v, got: %#v", c.ExpInput, in)
			}
		})
	}
}

func TestParseIntoNil(t *testing.T) {
	h := handling.NewH(
		encoding.NewStack(encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
	)

	r, _ := http.NewRequest("GET", "", nil)

	err := h.Parse(r, nil)
	if err != nil {
		t.Fatal("parsing into nil should be no-op")
	}
}
