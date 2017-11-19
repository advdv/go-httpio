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
	"github.com/advanderveer/go-httpio/encoding"
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

type testInput struct {
	Name     string `json:"json-name" schema:"form-name" validate:"ascii"`
	Position string `json:"position" schema:"position"`
}

type testOutput struct {
	Result string `json:"result,omitempty"`
}

func TestClientUsage(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Method    string
		Path      string
		Hdr       http.Header
		Ctrl      *httpio.Ctrl
		Input     *testInput
		Output    *testOutput
		ExpErr    error
		ExpOutput *testOutput
		Impl      func(context.Context, *testInput) (*testOutput, error)
	}{
		{
			Name:      "nil output",
			Method:    http.MethodGet,
			Output:    &testOutput{},
			ExpOutput: &testOutput{},
			Path:      "",
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}),
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
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
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}),
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return nil, errors.New("foo")
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				in := &testInput{}
				if render, valid := c.Ctrl.Handle(w, r, in); valid {
					render(c.Impl(r.Context(), in))
				}
			}))
			defer ts.Close()

			client, err := httpio.NewClient(ts.Client(), ts.URL, &encoding.JSON{})
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
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}),
			ExpBody:   `{}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return &testOutput{}, nil
			},
		},
		{
			Name:      "GET an nil output struct",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}),
			ExpBody:   `null` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return nil, nil
			},
		},

		{
			Name:      "GET return nil and error",
			Method:    http.MethodGet,
			Path:      "",
			Body:      nil,
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}),
			ExpBody:   `{"message":"foo"}` + "\n",
			ExpStatus: http.StatusInternalServerError,
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return nil, errors.New("foo")
			},
		},
		{
			Name:      "POST with query and json body",
			Method:    http.MethodPost,
			Path:      "?form-name=bar",
			Hdr:       http.Header{"Content-Type": []string{"application/json"}},
			Body:      strings.NewReader(`{"position": "director"}`),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"result":"bardirector"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
		{
			Name:      "POST with query and form body",
			Method:    http.MethodPost,
			Path:      "?form-name=bar",
			Hdr:       http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:      strings.NewReader("position=director"),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"result":"bardirector"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
		{
			Name:      "POST with query and json body that is invalid",
			Method:    http.MethodPost,
			Path:      "?form-name=bíar",
			Hdr:       http.Header{"Content-Type": []string{"application/json"}},
			Body:      strings.NewReader(`{"position": "director}`),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"message":"unexpected EOF"}` + "\n",
			ExpStatus: http.StatusBadRequest,
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			Impl: func(ctx context.Context, in *testInput) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(c.Method, c.Path, c.Body)
			r.Header = c.Hdr
			func(w http.ResponseWriter, r *http.Request) {
				in := &testInput{}
				if render, valid := c.Ctrl.Handle(w, r, in); valid {
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

func TestHandleInvalid(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Method    string
		Path      string
		Hdr       http.Header
		Body      io.Reader
		Ctrl      *httpio.Ctrl
		Impl      func(context.Context, *valTestInput, error) (*testOutput, error)
		ExpBody   string
		ExpStatus int
		ExpHdr    http.Header
	}{
		{
			Name:      "POST with query and form body",
			Method:    http.MethodPost,
			Path:      "?form-name=bar",
			Hdr:       http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:      strings.NewReader("position=director"),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"result":"bardirector"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *valTestInput, verr error) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
		{
			Name:      "POST with query and form body",
			Method:    http.MethodPost,
			Path:      "?form-name=invalid",
			Hdr:       http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			Body:      strings.NewReader("position=director"),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"result":"invaliddirectorinvalid name"}` + "\n",
			ExpStatus: http.StatusOK,
			ExpHdr:    http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Impl: func(ctx context.Context, in *valTestInput, verr error) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position + fmt.Sprint(verr)}, nil
			},
		},
		{
			Name:      "POST with query and json body that is invalid",
			Method:    http.MethodPost,
			Path:      "?form-name=bíar",
			Hdr:       http.Header{"Content-Type": []string{"application/json"}},
			Body:      strings.NewReader(`{"position": "director}`),
			Ctrl:      httpio.NewCtrl(&encoding.JSON{}, encoding.NewFormEncoding(schema.NewEncoder(), schema.NewDecoder())),
			ExpBody:   `{"message":"unexpected EOF"}` + "\n",
			ExpStatus: http.StatusBadRequest,
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			Impl: func(ctx context.Context, in *valTestInput, verr error) (*testOutput, error) {
				return &testOutput{Result: in.Name + in.Position}, nil
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(c.Method, c.Path, c.Body)
			r.Header = c.Hdr
			func(w http.ResponseWriter, r *http.Request) {
				in := &valTestInput{}
				if render, valid, verr := c.Ctrl.HandleInvalid(w, r, in); valid {
					render(c.Impl(r.Context(), in, verr))
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
