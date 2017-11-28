package httpio

import (
	"fmt"
	"mime"
	"net/http"
)

type decodeErr struct{ error }

func (e decodeErr) DecodeCause() bool { return true }

//IsDecodeErr can be used to
func IsDecodeErr(err error) bool {
	type isDecode interface {
		DecodeCause() bool
	}
	te, ok := err.(isDecode)
	return ok && te.DecodeCause()
}

//Ingress stack takes care of decoding incoming requests
type Ingress struct {
	egress   *Egress
	decoders DecoderList
	wares    []Transware
}

//NewIngress will setup the ingress stack, errors during parsing will be returned to using the egress stack.
func NewIngress(e *Egress, def DecoderFactory, others ...DecoderFactory) *Ingress {
	list := DecoderList{def}
	list = append(list, others...)
	return &Ingress{e, list, nil}
}

//Use will append the transware(s) to the egress render chain
func (i *Ingress) Use(wares ...Transware) {
	i.wares = append(i.wares, wares...)
}

func (i *Ingress) transformParse(next Transformer) Transformer {
	return TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error {
		if r.Body == nil || r.ContentLength == 0 {
			return next.Transform(a, r, w)
		}

		mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		e := i.decoders.Find(mt)
		if e == nil {
			return fmt.Errorf("httpio/ingress: unspported content type '%s'", mt)
		}

		dec := e.Decoder(r.Body)
		defer r.Body.Close()
		err := dec.Decode(a)
		if err != nil {
			return decodeErr{err} //tag with decode
		}

		return next.Transform(a, r, w)
	})
}

//Parse 'r' into 'in'
func (i *Ingress) Parse(r *http.Request, in interface{}) error {
	if in == nil {
		return nil //nothing to decode into
	}

	//in the case of ingress, our parse is always put in front of the middleware chain and the base is noop
	wares := append([]Transware{i.transformParse}, i.wares...)
	noop := TransFunc(func(a interface{}, r *http.Request, w http.ResponseWriter) error { return nil })
	chain := Chain(noop, wares...)
	err := chain.Transform(in, r, nil)
	if err != nil {
		return err
	}

	return nil
}

//Handle will parse request 'r' and decode it into 'in', it returns a renderfunction that is bound to response 'w'
func (i *Ingress) Handle(w http.ResponseWriter, r *http.Request, in interface{}) (fn RenderFunc, ok bool) {
	err := i.Parse(r, in)
	if err != nil {
		i.egress.MustRender(err, w, r)
		return nil, false
	}

	return func(out interface{}, err error) {
		if err != nil {
			i.egress.MustRender(err, w, r)
			return
		}

		i.egress.MustRender(out, w, r)
	}, true
}
