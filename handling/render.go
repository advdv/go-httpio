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

//MustRender will render 'v' onto 'w' but panics if this fails
func (h *H) MustRender(hdr http.Header, w http.ResponseWriter, v interface{}) {
	err := h.Render(hdr, w, v)
	if err != nil {
		panic(fmt.Sprintf("failed to render: %v", err))
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

	enc := e.Encoder(w)
	err = enc.Encode(v)
	if err != nil {
		return err
	}

	return nil
}
