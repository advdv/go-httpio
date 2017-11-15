package httpio

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

//ErrType is used to describe certain errors
type ErrType string

//IsErrType allows assertion of errors to be of a specific encoding type
func IsErrType(e error, etype ErrType) bool {
	if e == nil {
		return etype == ""
	}

	errh, ok := e.(herr)
	if !ok {
		return false
	}

	return errh.code == etype
}

const (
	//ErrParseForm indicates an error form encoded values
	ErrParseForm = ErrType("ERR_PARSE_FORM")

	//ErrDecodeForm is returned when decoding form values into structs fails
	ErrDecodeForm = ErrType("ERR_DECODE_FORM")

	//ErrUnsupportedContent is returned when the request's content is unsupported
	ErrUnsupportedContent = ErrType("ERR_UNSUPPORTED_CONTENT")

	//ErrDecodeBody is returned when decoding the body failed
	ErrDecodeBody = ErrType("ERR_DECODE_BODY")

	//ErrInputValidation is returned when validating the input failed
	ErrInputValidation = ErrType("ERR_INPUT_VALIDATION")

	//ErrNoDecoders is returned when no encoders were configured
	ErrNoDecoders = ErrType("ERR_NO_DECODERS")

	//ErrRenderOutput is returned when custom output rendering fails
	ErrRenderOutput = ErrType("ERR_RENDER_OUTPUT")

	//ErrRenderFatal is returned when custom output rendering panicks
	ErrRenderFatal = ErrType("ERR_RENDER_FATAL")
)

//IO provides methods for parsing http.Request into inputs and outputs into
//http.ResponseWriters
type IO struct {
	Decoders       map[string]func(io.Reader) Decoder
	Encoders       map[string]func(io.Writer) Encoder
	FormDecoder    FormDecoder
	InputValidator InputValidator
}

//NewJSON returns an http io setup that is just for json input and output
func NewJSON(fdec FormDecoder, val InputValidator) *IO {
	return &IO{
		Decoders: map[string]func(io.Reader) Decoder{
			"application/json": func(r io.Reader) Decoder { return json.NewDecoder(r) },
		},
		Encoders: map[string]func(io.Writer) Encoder{
			"application/json": func(w io.Writer) Encoder { return json.NewEncoder(w) },
		},
		FormDecoder:    fdec,
		InputValidator: val,
	}
}

//Decoder is used for non-form values such as JSON and XML.
type Decoder interface {
	Decode(v interface{}) error
}

//Encoder is used for non-form values such as JSON ans XML
type Encoder interface {
	Encode(v interface{}) error
}

//Err can be embedded into output structs to enable servers to
//send error messages over to the client as a string:
type Err string

//StatusCode implements the custom status interface, it returns 500 of
//there is an embedded error
func (err *Err) StatusCode() int {
	if err.UnboxErr() == nil {
		return -1
	}

	return http.StatusInternalServerError
}

//BoxErr implements a method that allows any type that embeds E
//to allow basic piggybacking of error values
func (err *Err) BoxErr(e error) {
	*err = Err(e.Error())
}

//UnboxErr returns nil if the embedding struct doesn't have an error value
//embedded or the actual error if it does
func (err *Err) UnboxErr() error {
	if err == nil || *err == "" {
		return nil
	}

	return errors.New(string(*err))
}

//BoundRender is returned by the parse function and is bound to the ResponseWriter it
//was called with and the httpio.IO it was called on.
type BoundRender func(out interface{}, appErr error)

//InputValidator can be implemented to provide validation of request input structs
//Incidentally, this is implemented by `github.com/go-playground/validator`
type InputValidator interface {
	StructCtx(ctx context.Context, input interface{}) (err error)
}

//FormDecoder can be implemented to provide decoding of form values into input structs
//Incidentally, this interface is implemented immediately by `github.com/gorilla/schema`
type FormDecoder interface {
	Decode(dst interface{}, src map[string][]string) error
}
