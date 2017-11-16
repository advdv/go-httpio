package handling_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

type renderMyself string

func (r renderMyself) Render(hdr http.Header, w http.ResponseWriter) error {
	fmt.Fprintf(w, "hello, %s: %v", r, hdr)
	return nil
}

func TestRender(t *testing.T) {
	for _, c := range []struct {
		Name       string
		Handling   *handling.H
		Hdr        http.Header
		Value      interface{}
		ExpContent string
		ExpError   error
	}{
		{
			Name: "json rendering of nil",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{}},
			Value:      nil,
			ExpContent: `null` + "\n",
			ExpError:   nil,
		},
		{
			Name: "custom rendering on type",
			Handling: handling.NewH(encoding.NewStack(
				&encoding.JSON{},
			)),
			Hdr:        http.Header{"Content-Type": []string{"application/json"}},
			Value:      renderMyself("world"),
			ExpContent: `hello, world: map[Content-Type:[application/json]]`,
			ExpError:   nil,
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			hdr := c.Hdr
			rec := httptest.NewRecorder()
			err := c.Handling.Render(hdr, rec, c.Value)
			if err != c.ExpError {
				t.Fatalf("expected err '%v' to be '%v'", err, c.ExpError)
			}

			if rec.Body.String() != c.ExpContent {
				t.Fatalf("expected output '%s' to be '%s'", rec.Body.String(), c.ExpContent)
			}
		})
	}

}
