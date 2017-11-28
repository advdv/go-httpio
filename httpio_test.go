package httpio_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/gorilla/schema"
)

type valTestInput struct {
	Name     string `json:"json-name" schema:"form-name" validate:"ascii"`
	Position string `json:"position" schema:"position"`
}

func (i *valTestInput) Validate() error {
	if i.Name == "invalid" {
		return errors.New("invalid name")
	}

	return nil
}

type testInput2 struct {
	Name     string `json:"json-name" schema:"form-name" validate:"ascii"`
	Position string `json:"position" schema:"position"`
}

type testOutput struct {
	Result string `json:"result,omitempty"`
}

var stdQueryWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		vals := r.URL.Query()
		if len(vals) > 0 {
			dec := schema.NewDecoder()
			err := dec.Decode(a, vals)
			if err != nil {
				return err
			}
		}

		return next.Transform(a, r, w)
	})
}

var stdErrWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		if err, ok := a.(error); ok {
			a = struct {
				Message string `json:"message"` //give the error response shape
			}{err.Error()}

			w.Header().Set("X-Has-Handling-Error", "1")
			errStatus := http.StatusInternalServerError
			if httpio.IsDecodeErr(err) {
				errStatus = http.StatusBadRequest
			}

			return next.Transform(a, r.WithContext(httpio.WithStatus(r.Context(), errStatus)), w) //set a status code
		}

		return next.Transform(a, r, w)
	})
}

func newStdIO() *httpio.Ingress {
	j := &httpio.JSON{}
	egress := httpio.NewEgress(j)
	egress.Use(stdErrWare)
	ingress := httpio.NewIngress(egress, j, httpio.NewFormDecoding(schema.NewDecoder()))
	ingress.Use(stdQueryWare)
	return ingress
}

func TestClientUsage(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Method    string
		Path      string
		Hdr       http.Header
		Ingress   *httpio.Ingress
		Input     *testInput2
		Output    *testOutput
		ExpErr    error
		ExpOutput *testOutput
		Impl      func(context.Context, *testInput2) (*testOutput, error)
	}{
		{
			Name:      "nil output",
			Method:    http.MethodGet,
			Output:    &testOutput{},
			ExpOutput: &testOutput{},
			Path:      "",
			Ingress:   newStdIO(),
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return nil, nil
			},
		},
		{
			Name:      "return error",
			Method:    http.MethodGet,
			Output:    &testOutput{},
			ExpOutput: &testOutput{},
			ExpErr:    errors.New("foo"),
			Path:      "",
			Ingress:   newStdIO(),
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return nil, errors.New("foo")
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				in := &testInput2{}
				if render, valid := c.Ingress.Handle(w, r, in); valid {
					render(c.Impl(r.Context(), in))
				}
			}))
			defer ts.Close()

			client, err := httpio.NewClient(ts.Client(), ts.URL, &httpio.JSON{}, &httpio.JSON{})
			if err != nil {
				t.Fatal("failed to create client:", err)
			}

			err = client.Request(context.Background(), c.Method, c.Path, c.Hdr, c.Input, c.Output)
			if fmt.Sprint(err) != fmt.Sprint(c.ExpErr) {
				t.Fatalf("expected err '%v', got: '%v'", c.ExpErr, err)
			}

			if !reflect.DeepEqual(c.Output, c.ExpOutput) {
				t.Fatalf("expected output '%v', got: %v", c.ExpOutput, c.Output)
			}
		})
	}
}

func TestUsageWithoutClient(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Method    string
		Path      string
		Hdr       http.Header
		Body      io.Reader
		Ingress   *httpio.Ingress
		Impl      func(context.Context, *testInput2) (*testOutput, error)
		ExpBody   string
		ExpStatus int
		ExpHdr    http.Header
	}{
		{
			Name:      "GET an empty output struct",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ingress:   newStdIO(),
			ExpBody:   `{}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return &testOutput{}, nil
			},
		},
		{
			Name:      "GET an nil output struct",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ingress:   newStdIO(),
			ExpBody:   `null` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return nil, nil
			},
		},
		{
			Name:      "GET return nil and error",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ingress:   newStdIO(),
			ExpBody:   `{"message":"foo"}` + "\n",
			ExpStatus: http.StatusInternalServerError,
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return nil, errors.New("foo")
			},
		},
		{
			Name:      "POST with query and json body",
			Method:    http.MethodPost,
			Path:      "?form-name=bar",
			Hdr:       http.Header{"Content-Type": []string{"application/json"}},
			Body:      strings.NewReader(`{"position": "director"}`),
			Ingress:   newStdIO(),
			ExpBody:   `{"result":"bardirector"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
		{
			Name:      "POST with query and form body",
			Method:    http.MethodPost,
			Path:      "?form-name=bar",
			Hdr:       http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:      strings.NewReader("position=director"),
			Ingress:   newStdIO(),
			ExpBody:   `{"result":"bardirector"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
		{
			Name:      "POST with query and json body that is invalid",
			Method:    http.MethodPost,
			Path:      "?form-name=bíar",
			Hdr:       http.Header{"Content-Type": []string{"application/json"}},
			Body:      strings.NewReader(`{"position": "director}`),
			Ingress:   newStdIO(),
			ExpBody:   `{"message":"unexpected EOF"}` + "\n",
			ExpStatus: http.StatusBadRequest,
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			Impl: func(ctx context.Context, in *testInput2) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(c.Method, c.Path, c.Body)
			r.Header = c.Hdr
			func(w http.ResponseWriter, r *http.Request) {
				in := &testInput2{}
				if render, valid := c.Ingress.Handle(w, r, in); valid {
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
