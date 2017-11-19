package handling

import (
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/advanderveer/go-httpio/encoding"
)

//Validate can be implemeted by inputs that can validate themselves
type Validate interface {
	Validate() error
}

func (h *H) parseForm(r *http.Request, v interface{}) error {
	fenc := h.encs.Find(encoding.MediaTypeForm)
	if fenc == nil {
		return nil
	}

	err := r.ParseForm()
	if err != nil {
		return err //invalid format or escaping
	}

	q := strings.NewReader(r.Form.Encode())
	dec := fenc.Decoder(q)
	err = dec.Decode(v)
	if err != nil {
		return err //unsupported types etc
	}

	if h.Validator != nil {
		err := h.Validator.Validate(v)
		if err != nil {
			return err //validator for all values bassing by
		}
	}

	if vv, ok := v.(Validate); ok {
		err := vv.Validate()
		if err != nil {
			return err //validate specific values
		}
	}

	return nil
}

func (h *H) parseContent(r *http.Request, v interface{}) (err error) {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}

	mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	e := h.encs.Find(mt)
	if e == nil {
		return errors.New("unspported content type")
	}

	dec := e.Decoder(r.Body)
	defer r.Body.Close()
	err = dec.Decode(v)
	if err != nil {
		return err
	}

	return nil
}

//Parse will atempt to parse the http request 'r' into 'v'
func (h *H) Parse(r *http.Request, v interface{}) error {
	err := h.parseForm(r, v)
	if err != nil {
		return err
	}

	err = h.parseContent(r, v)
	if err != nil {
		return err
	}

	return nil
}
