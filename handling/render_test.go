package handling_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

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
	} {
		t.Run(c.Name, func(t *testing.T) {
			hdr := http.Header{}
			buf := bytes.NewBuffer(nil)
			err := c.Handling.Render(hdr, buf, c.Value)
			if err != c.ExpError {
				t.Fatalf("expected err '%v' to be '%v'", err, c.ExpError)
			}

			if buf.String() != c.ExpContent {
				t.Fatalf("expected output '%s' to be '%s'", buf.String(), c.ExpContent)
			}
		})
	}

}
