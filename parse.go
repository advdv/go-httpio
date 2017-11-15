package httpio

import (
	"fmt"
	"mime"
	"net/http"
)

type herr struct {
	code    ErrType
	message string
	cause   error
	status  int
}

//Error implements the error interface
func (pe herr) Error() string {
	return fmt.Sprintf("%s: %s", pe.code, pe.message)
}

//parse will attemp to parse request 'r' into input 'in' and return an error when it fails
func (hio *IO) parse(r *http.Request, in interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return herr{ErrParseForm, "failed to parse query or form encoded parameters", err, http.StatusBadRequest}
	}

	if len(r.Form) > 0 && hio.FormDecoder != nil {
		err = hio.FormDecoder.Decode(in, r.Form)
		if err != nil {
			return herr{ErrDecodeForm, "failed to decode query or form encoded parameters", err, http.StatusBadRequest}
		}
	}

	mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if r.Body != nil && r.ContentLength != 0 {
		if hio.Decoders == nil || len(hio.Decoders) < 1 {
			return herr{ErrNoDecoders, "failed to decode request body", nil, http.StatusInternalServerError}
		}

		decf, ok := hio.Decoders[mt]
		if !ok {
			return herr{ErrUnsupportedContent, "content type of request body is unsupported", fmt.Errorf(mt), http.StatusUnsupportedMediaType}
		}

		dec := decf(r.Body)
		err = dec.Decode(in)
		if err != nil {
			return herr{ErrDecodeBody, "failed to decode request body", err, http.StatusBadRequest}
		}
	}

	if hio.InputValidator != nil {
		err = hio.InputValidator.StructCtx(r.Context(), in)
		if err != nil {
			return herr{ErrInputValidation, "provided input fields are invalid or missing", err, http.StatusBadRequest}
		}
	}

	return nil
}

func (hio *IO) boxErrOrRender(w http.ResponseWriter, r *http.Request, out interface{}, err error) {
	if box, ok := out.(interface {
		BoxErr(e error)
	}); ok {
		box.BoxErr(err)
		hio.Render(w, r, out) //we can sent the error as part of the output message
		return
	}

	hio.Render(w, r, err) //we just render the error ourselves
}

//Parse will parse request 'r' and if it succeeds returns a render func that is bound to render to 'w'
func (hio *IO) Parse(w http.ResponseWriter, r *http.Request, in, out interface{}) (BoundRender, bool) {
	if err := hio.parse(r, in); err != nil {
		hio.boxErrOrRender(w, r, out, err)
		return nil, false
	}

	return func(ranOut interface{}, appErr error) {
		if appErr != nil {
			//we box in the empty output struct, the one that was ran might contain sensitive values
			//that the user might have though to be discarded upon returning an error from the implementation
			hio.boxErrOrRender(w, r, out, appErr)
			return
		}

		hio.Render(w, r, ranOut)
	}, true
}
