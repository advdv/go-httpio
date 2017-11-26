package httpio

import (
	"net/http"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
)

//Ctrl allows for handling HTTP by defining supported encoding schemes
type Ctrl struct {
	H *handling.H
}

//NewCtrl sets up a controller that handles the provided encoding schemes
func NewCtrl(def encoding.Encoding, other ...encoding.Encoding) *Ctrl {
	return &Ctrl{
		H: handling.NewH(encoding.NewStack(def, other...)),
	}
}

//RenderFunc is bound to an request but renders once called
type RenderFunc func(interface{}, error)

//SetErrorHandler provides a place to centrally handle implementation errors
func (c *Ctrl) SetErrorHandler(errh handling.ErrHandler) {
	c.H.ErrHandler = errh
}

func (c *Ctrl) bind(w http.ResponseWriter, r *http.Request) RenderFunc {
	return func(out interface{}, err error) {
		if err != nil {
			c.H.MustRender(w, r, err)
			return
		}

		c.H.MustRender(w, r, out)
	}
}

//Handle will parse request 'r' into input 'in' and renders when 'f' is called. If
//valid is false it means decoding or validation failed the error is instead immediately
//rendered onto 'w', no further attempt at rendering should be done at this point.
func (c *Ctrl) Handle(w http.ResponseWriter, r *http.Request, in interface{}) (f RenderFunc, valid bool) {
	err := c.H.Parse(r, in)
	if err != nil {
		c.H.MustRender(w, r, err)
		return nil, false
	}

	return c.bind(w, r), true
}
