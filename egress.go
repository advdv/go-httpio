package httpio

import (
	"context"
	"fmt"
	"net/http"
)

type contextValue string

var (
	contextValueStatusCode = contextValue("status_code")
)

//StatusValue returns a specific status stored in the (request) context, returns 0 if its not specified
func StatusValue(ctx context.Context) (code int) {
	code, _ = ctx.Value(contextValueStatusCode).(int)
	return
}

//WithStatus will write a status to the (request)Context for transware down the chain
func WithStatus(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, contextValueStatusCode, code)
}

//RenderFunc is bound to an request but renders once called
type RenderFunc func(interface{}, error)

//Egress takes care of encoding outgoing responses
type Egress struct {
	encoders EncoderList
	wares    []Transware
}

//NewEgress uses the provided encoder factories to setup encoding
func NewEgress(def EncoderFactory, others ...EncoderFactory) *Egress {
	list := EncoderList{def}
	list = append(list, others...)
	return &Egress{list, nil}
}

//MustRender will render 'out' onto 'w' if this fails it will attemp to render the error. If this
//fails, it panics.
func (e *Egress) MustRender(out interface{}, w http.ResponseWriter, r *http.Request) {
	err := e.Render(out, w, r)
	if err != nil {
		err = e.Render(err, w, r)
		if err != nil {
			panic("failed to render: " + err.Error())
		}
	}
}

func (e *Egress) encode(a interface{}, r *http.Request, w http.ResponseWriter) error {
	status := StatusValue(r.Context())
	if status == 0 {
		status = http.StatusOK
	}

	mt := negotiateContentType(r.Header, e.encoders.Supported(), e.encoders.Default().MimeType())
	encf := e.encoders.Find(mt)
	if encf == nil {
		return fmt.Errorf("httpio/egress: no encoder for media type '%s'", mt)
	}

	enc := encf.Encoder(w)
	w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mt))
	w.WriteHeader(status)
	err := enc.Encode(a)
	if err != nil {
		return err
	}

	return nil
}

//Use will append the transware(s) to the egress render chain
func (e *Egress) Use(wares ...Transware) {
	e.wares = append(e.wares, wares...)
}

//Render will take value 'v' and encode it onto response 'w' in context of request 'r'
func (e *Egress) Render(out interface{}, w http.ResponseWriter, r *http.Request) (err error) {
	chain := Chain(TransFunc(e.encode), e.wares...)
	err = chain.Transform(out, r, w)
	if err != nil {
		return err
	}

	return nil
}
