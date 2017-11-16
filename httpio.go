package httpio

import (
	"net/http"

	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/errbox"
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

//Handle will parse request 'r' into input 'in' and bind 'out' to be render func 'f' is called. If
//valid is false the error is already rendered onto 'w', no further attempt at rendering should be
//done at this point.
func (c *Ctrl) Handle(w http.ResponseWriter, r *http.Request, in, errOut interface{}) (f RenderFunc, valid bool) {
	err := c.H.Parse(r, in)
	if err != nil {
		errOut, _ = errbox.Box(errOut, err)
		c.H.MustRender(r.Header, w, errOut)
		return nil, false
	}

	//@TODO validate, then box (in errOut) or render error

	return func(out interface{}, err error) {
		out, _ = errbox.Box(out, err)
		c.H.MustRender(r.Header, w, out)
	}, true
}
