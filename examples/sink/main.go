package main

import (
	"context"
	"log"
	"net/http"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/advanderveer/go-httpio/encoding"
	"github.com/monoculum/formam"
	"github.com/roobre/gorilla-mux"
)

type myService struct{}

func (s *myService) Echo(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	return in, nil
}

type formDecoder struct {
	*formam.Decoder
}

func (fd *formDecoder) Decode(dst interface{}, vs map[string][]string) error {
	return fd.Decoder.Decode(vs, dst)
}

func main() {
	r := mux.NewRouter() //route requests to handlers

	ctrl := httpio.NewCtrl(
		&encoding.JSON{},
		encoding.NewFormEncoding(nil, &formDecoder{formam.NewDecoder(nil)}),
	) //request parsing and response rendering (including errors)

	svc := &myService{} //services implement business logic

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		input := &map[string]interface{}{}
		if render, ok := ctrl.Handle(w, r, input); ok {
			render(svc.Echo(r.Context(), *input))
		}
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
