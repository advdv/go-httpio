package httpio

import (
	"net/http"
)

//Render will render output 'out' onto response writer 'w' using type assertions and content negotiation
//based on request 'r'
func (hio *IO) Render(w http.ResponseWriter, r *http.Request, out interface{}) {
	status := http.StatusOK
	if vv, ok := out.(interface {
		StatusCode() int //this interface allows rendered values to just customize the status code
	}); ok {
		if s := vv.StatusCode(); s > 0 {
			status = s
		}
	}

	rendered := false
	if vv, ok := out.(interface {
		Render(wr http.ResponseWriter) error //custom rendering for this output
	}); ok {
		func() {
			defer func() {
				if r := recover(); r != nil {
					out = herr{ErrRenderFatal, "fatal error during output rendering", nil, http.StatusInternalServerError}
				}
			}()

			err := vv.Render(w)
			if err != nil {
				out = herr{ErrRenderOutput, "output rendering failed", err, http.StatusInternalServerError}
			} else {
				rendered = true
			}
		}()
		if rendered {
			return
		}
	}

	if vv, ok := out.(error); ok {
		status = http.StatusInternalServerError
		out = map[string]string{
			"message": vv.Error(),
		}
	}

	supported := []string{}
	for mime := range hio.Encoders {
		supported = append(supported, mime)
	}

	if len(supported) < 1 {
		http.Error(w, "no encoder supported", http.StatusInternalServerError)
		return
	}

	var err error
	ct := negotiateContentType(r, supported, supported[0])
	enc := hio.Encoders[ct](w)

	w.Header().Set("Content-Type", ct+"; charset=utf-8")
	w.WriteHeader(status)
	err = enc.Encode(out)
	if err != nil {
		http.Error(w, "failed to encode response body", http.StatusInternalServerError)
	}
}
