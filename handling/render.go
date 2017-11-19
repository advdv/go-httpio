package handling

import (
	"errors"
	"fmt"
	"net/http"
)

//Renderable indicates that a type is able to render itself
type Renderable interface {
	Render(r *http.Request, w http.ResponseWriter) error
}

//Statuser allows a type to provide its own status code
type Statuser interface {
	StatusCode() int
}

//MustRender will render 'v' onto 'w' if this fails it will attemp to render the error. If this
//this fails, it panics
func (h *H) MustRender(w http.ResponseWriter, r *http.Request, v interface{}) {
	err := h.Render(w, r, v)
	if err != nil {
		err = h.Render(w, r, err)
		if err != nil {
			panic(fmt.Sprintf("failed to render: %v", err))
		}
	}
}

//Render 'v' onto 'w' by takikng into account preferences in 'hdr'
func (h *H) Render(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if errv, ok := v.(error); ok {
		v = h.ErrHandler(r.Context(), errv, w.Header())
	}

	if renderable, ok := v.(Renderable); ok {
		err := renderable.Render(r, w)
		if err != nil {
			return err
		}

		return nil
	}

	var status int
	if statuser, ok := v.(Statuser); ok {
		status = statuser.StatusCode()
	} else {
		status = http.StatusOK
	}

	offers := h.encs.Supported()
	doffer, err := h.encs.Default()
	if err != nil {
		return err
	}

	mt := negotiateContentType(r.Header, offers, doffer.MimeType())
	e := h.encs.Find(mt)
	if e == nil {
		return errors.New("no suitable encoding found")
	}

	w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mt))
	w.WriteHeader(status)

	enc := e.Encoder(w)
	err = enc.Encode(v)
	if err != nil {
		return err
	}

	return nil
}
