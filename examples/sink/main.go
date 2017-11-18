package main

import (
	"log"
	"net/http"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/advanderveer/go-httpio/encoding"
	"github.com/gorilla/mux"
	"github.com/monoculum/formam"
	"gopkg.in/go-playground/validator.v9"
)

type inputValidator struct {
	v *validator.Validate
}

func (ip *inputValidator) Validate(v interface{}) error {
	return ip.v.Struct(v)
}

type formDecoder struct {
	*formam.Decoder
}

func (fd *formDecoder) Decode(dst interface{}, vs map[string][]string) error {
	return fd.Decoder.Decode(vs, dst)
}

func main() {
	r := mux.NewRouter()
	ctrl := httpio.NewCtrl(
		&inputValidator{validator.New()},
		&encoding.JSON{},
		encoding.NewFormEncoding(nil, &formDecoder{formam.NewDecoder(nil)}),
	)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		in := &map[string]interface{}{}
		out := &map[string]interface{}{}
		if render, ok := ctrl.Handle(w, r, in, out); ok {
			render(func() (*map[string]interface{}, error) {

				return in, nil
			}())
		}
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
