package handling_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

type renderMyselfFail string

func (rmf renderMyselfFail) Render(r *http.Request, w http.ResponseWriter) error {
	return errors.New("my rendering error")
}

type renderMyself string

func (rmf renderMyself) Render(r *http.Request, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "foo/bar")
	w.WriteHeader(900)
	fmt.Fprintf(w, "hello, %s: %v", rmf, r.Header)

	return nil
}

type overwriteDelegate func(v interface{}, r *http.Request) (interface{}, error)

func (f overwriteDelegate) PreRender(v interface{}, r *http.Request) (interface{}, error) {
	return f(v, r)
}

type statusCodeError struct{ error }

func (e statusCodeError) StatusCode() int { return http.StatusBadRequest }

type renderStatus struct{}

func (r *renderStatus) StatusCode() int { return 1000 }

type notFoundErr string

func (r notFoundErr) Error() string { return string(r) }

func (r notFoundErr) StatusCode() int { return http.StatusNotFound }

func TestRender(t *testing.T) {
	for _, c := range []struct {
		Name        string
		Must        bool
		Handling    *handling.H
		Hdr         http.Header
		Value       interface{}
		ExpContent  string
		EncDelegate handling.EncodingDelegate
		ExpHdr      http.Header
		ExpError    error
		ExpStatus   int
	}{
		{
			Name: "json rendering of nil",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{}},
			Value:      nil,
			ExpContent: `null` + "\n",
			ExpHdr:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			ExpError:   nil,
			ExpStatus:  200,
		},
		{
			Name: "custom rendering on type",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      renderMyself("world"),
			ExpContent: `hello, world: map[Content-Type:[application/json]]`,
			ExpHdr:     http.Header{"Content-Type": []string{"foo/bar"}},
			ExpError:   nil,
			ExpStatus:  900,
		},
		{
			Name: "custom rendering on type with encoding delegate",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:   http.Header{"Content-Type": []string{"application/json"}},
			Value: renderMyself("world"),
			EncDelegate: overwriteDelegate(func(v interface{}, r *http.Request) (interface{}, error) {
				if _, ok := v.(renderMyself); ok {
					return renderMyself("overwritten"), nil
				}

				return v, nil
			}),
			ExpContent: `hello, overwritten: map[Content-Type:[application/json]]`,
			ExpHdr:     http.Header{"Content-Type": []string{"foo/bar"}},
			ExpError:   nil,
			ExpStatus:  900,
		},
		{
			Name: "custom status on type",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      &renderStatus{},
			ExpContent: `{}` + "\n",
			ExpHdr:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			ExpError:   nil,
			ExpStatus:  1000,
		},
		{
			Name: "render generic error",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      errors.New("foo"),
			ExpContent: `{"message":"foo"}` + "\n",
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			ExpError:  nil,
			ExpStatus: http.StatusInternalServerError,
		},
		{
			Name: "render custom error",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      statusCodeError{errors.New("foo")},
			ExpContent: `{"message":"foo"}` + "\n",
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			ExpError:  nil,
			ExpStatus: http.StatusBadRequest,
		},
		{
			Name: "custom rendering that fails on type",
			Must: true,
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      renderMyselfFail("world"),
			ExpContent: `{"message":"my rendering error"}` + "\n",
			ExpHdr: http.Header{
				"Content-Type":         []string{"application/json; charset=utf-8"},
				"X-Has-Handling-Error": []string{"1"},
			},
			ExpError:  nil,
			ExpStatus: http.StatusInternalServerError,
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			req.Header = c.Hdr

			if c.EncDelegate != nil {
				c.Handling.EncodingDelegate = c.EncDelegate
			}

			rec := httptest.NewRecorder()
			if c.Must {
				c.Handling.MustRender(rec, req, c.Value)
			} else {
				err := c.Handling.Render(rec, req, c.Value)
				if err != c.ExpError {
					t.Fatalf("expected err '%v' to be '%v'", err, c.ExpError)
				}
			}

			if rec.Code != c.ExpStatus {
				t.Fatalf("expected status '%d', got: %d", c.ExpStatus, rec.Code)
			}

			if !reflect.DeepEqual(rec.Header(), c.ExpHdr) {
				t.Fatalf("expected resp hdr '%v', got: %v", c.ExpHdr, rec.Header())
			}

			if rec.Body.String() != c.ExpContent {
				t.Fatalf("expected output '%s' to be '%s'", rec.Body.String(), c.ExpContent)
			}
		})
	}
}
