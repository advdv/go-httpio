package httpio_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	httpio "github.com/advanderveer/go-httpio"
)

//usecase: vanille
var nopWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		return next.Transform(a, r, w)
	})
}

//usecase: an error handler
var errWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		if err, ok := a.(error); ok {
			a = struct {
				Message string `json:"message"`
			}{err.Error()}
		}

		return next.Transform(a, r, w)
	})
}

type rendersItself string

func (r rendersItself) Render(w http.ResponseWriter) {
	fmt.Fprintf(w, `{"bar": "%s"}`+"\n", r)
}

//usecase: an output struct can render itself
var earlyWriteWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		if r, ok := a.(rendersItself); ok {
			r.Render(w)
			return nil
		}

		return next.Transform(a, r, w)
	})
}

//usecase: one middleware returns an error
var earlyReturnErrWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		return errors.New("early error")
	})
}

//usecase: middleware determines status code
var statusWare = func(next httpio.Transformer) httpio.Transformer {
	return httpio.TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		return next.Transform(a, r.WithContext(httpio.WithStatus(r.Context(), 900)), w)
	})
}

func TestJustBaseChain(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Wares     []httpio.Transware
		Value     interface{}
		ExpErr    error
		ExpStatus int
		ExpBody   string
	}{
		{
			Name:      "null without middleware",
			ExpStatus: 200,
			ExpBody:   fmt.Sprintln(`null`),
		},
		{
			Name:      "map with no-op middleware",
			Value:     map[string]string{"foo": "bar"},
			Wares:     []httpio.Transware{nopWare},
			ExpStatus: 200,
			ExpBody:   fmt.Sprintln(`{"foo":"bar"}`),
		},
		{
			Name:      "error with error middleware",
			Value:     errors.New("some error"),
			Wares:     []httpio.Transware{errWare},
			ExpStatus: 200,
			ExpBody:   fmt.Sprintln(`{"message":"some error"}`),
		},
		{
			Name:      "early renderer middleware",
			Value:     rendersItself("i render myself"),
			Wares:     []httpio.Transware{earlyWriteWare},
			ExpStatus: 200,
			ExpBody:   fmt.Sprintln(`{"bar": "i render myself"}`),
		},
		{
			Name:      "early error middleware",
			Value:     map[string]string{"foo": "bar"},
			Wares:     []httpio.Transware{earlyReturnErrWare},
			ExpErr:    errors.New("early error"),
			ExpStatus: 200,
		},
		{
			Name:      "custom status ware",
			Value:     map[string]string{"foo": "bar"},
			Wares:     []httpio.Transware{statusWare},
			ExpStatus: 900,
			ExpBody:   fmt.Sprintln(`{"foo":"bar"}`),
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			e := httpio.NewEgress(&httpio.JSON{})
			e.Use(c.Wares...)

			r, _ := http.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			err := e.Render(c.Value, w, r)
			if fmt.Sprint(err) != fmt.Sprint(c.ExpErr) {
				t.Fatalf("expected error '%s', got: '%s'", c.ExpErr, err)
			}

			if w.Code != c.ExpStatus {
				t.Fatalf("expected status %d, got: %d", c.ExpStatus, w.Code)
			}

			if w.Body.String() != c.ExpBody {
				t.Fatalf("expected body %s, got: %s", c.ExpBody, w.Body.String())
			}
		})
	}

}
