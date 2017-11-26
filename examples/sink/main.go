package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	httpio "github.com/advanderveer/go-httpio"
	"github.com/advanderveer/go-httpio/encoding"
	"github.com/advanderveer/go-httpio/handling"
	"github.com/monoculum/formam"
	"github.com/roobre/gorilla-mux"
)

type myInvalidErr interface {
	IsMyInvalidMessage()
}

//errInvalidMessage is our custom error type
type errInvalidMessage struct{ error }

func (eim errInvalidMessage) IsMyInvalidMessage() {}

type myMessage struct {
	A string `json:"a" formam:"a" validate:"min=4"`
}

//on struct validation example
func (msg *myMessage) Validate() error {
	if msg.A == "invalid" {
		return errInvalidMessage{errors.New("my invalid error")}
	}
	return nil
}

type myPage struct {
	t *template.Template
	*myMessage
}

//implements the renderable interface for custom rendering
func (p *myPage) Render(r *http.Request, w http.ResponseWriter) error {
	return p.t.Execute(w, p)
}

type myService struct{}

func (s *myService) Echo(ctx context.Context, in *myMessage) (*myMessage, error) {
	return in, nil
}

func (s *myService) Page(ctx context.Context, in *myMessage) (*myPage, error) {
	return &myPage{
		myMessage: in,
		t:         template.Must(template.New("page").Parse(`hello, {{.A}}` + "\n")),
	}, nil
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

	ctrl.SetErrorHandler(func(ctx context.Context, err error, whdr http.Header) interface{} {
		cause := errors.Cause(err) //parse errors that arrive happen to implement the cause interface
		if _, ok := cause.(myInvalidErr); ok {
			err = errors.New(strings.ToUpper(err.Error())) //customize our errors before sending back
		}

		return handling.HeaderErrHandling(ctx, err, whdr) //call original if we want
	})

	svc := &myService{} //services implement business logic

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		input := &myMessage{}
		if render, ok := ctrl.Handle(w, r, input); ok {
			render(svc.Echo(r.Context(), input))
		}
	}).Methods("POST")

	r.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		input := &myMessage{}
		if render, ok := ctrl.Handle(w, r, input); ok {
			render(svc.Page(r.Context(), input))
		}
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
