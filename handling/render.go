package handling

import (
	"errors"
	"fmt"
	"net/http"
)

//Renderable indicates that a type is able to render itself
type Renderable interface {
	Render(hdr http.Header, w http.ResponseWriter) error
}

//Statuser allows a type to provide its own status code
type Statuser interface {
	StatusCode() int
}

//MustRender will render 'v' onto 'w' if this fails it will attemp to render the error. If this
//this fails, it panics
func (h *H) MustRender(hdr http.Header, w http.ResponseWriter, v interface{}) {
	err := h.Render(hdr, w, v)
	if err != nil {
		err = h.Render(hdr, w, err)
		if err != nil {
			panic(fmt.Sprintf("failed to render: %v", err))
		}
	}
}

//Render 'v' onto 'w' by takikng into account preferences in 'hdr'
func (h *H) Render(hdr http.Header, w http.ResponseWriter, v interface{}) error {
	if renderable, ok := v.(Renderable); ok {
		err := renderable.Render(hdr, w)
		if err != nil {
			return err
		}

		return nil
	}

	status := -1
	if statuser, ok := v.(Statuser); ok {
		status = statuser.StatusCode()
	}

	if errv, ok := v.(error); ok {
		if status == -1 {
			status = http.StatusInternalServerError
		}

		w.Header().Set(HeaderHandlingError, "1")
		v = Err{Message: errv.Error()}
	}

	if status == -1 {
		status = http.StatusOK
	}

	offers := h.encs.Supported()
	doffer, err := h.encs.Default()
	if err != nil {
		return err
	}

	mt := negotiateContentType(hdr, offers, doffer.MimeType())
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
