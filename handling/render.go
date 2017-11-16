package handling

import (
	"errors"
	"io"
	"net/http"
)

//Render 'v' onto 'w' by takikng into account preferences in 'hdr'
func (h *H) Render(hdr http.Header, w io.Writer, v interface{}) error {
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
