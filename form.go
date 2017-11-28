package httpio

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
)

var (
	//MediaTypeForm identifies form content
	MediaTypeForm = "application/x-www-form-urlencoded"
)

//FormDecodeProvider can be implemented to provide decoding of form maps into structs
//Incidentally, this interface is implemented immediately by `github.com/gorilla/schema`
type FormDecodeProvider interface {
	Decode(dst interface{}, src map[string][]string) error
}

//FormEncodeProvider can be implemented to provide encoding of form maps from structs
//Incidentally, this interface is implemented immediately by `github.com/gorilla/schema`
type FormEncodeProvider interface {
	Encode(src interface{}, dst map[string][]string) error
}

//FormEncoder uses the form encoding provider to implement the Encoding interface
type FormEncoder struct {
	enc FormEncodeProvider
	w   io.Writer
}

//Encode the value v into the encoder writer
func (e *FormEncoder) Encode(v interface{}) error {
	if e.enc == nil {
		return errors.New("no form encoder configured")
	}

	vals := url.Values{}
	err := e.enc.Encode(v, vals)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(e.w, vals.Encode())
	return err
}

//FormDecoder uses the form encoding provider to implement the Encoding interface
type FormDecoder struct {
	dec FormDecodeProvider
	r   io.Reader
}

//Decode into v from the reader
func (e *FormDecoder) Decode(v interface{}) error {
	if e.dec == nil {
		return errors.New("no form decoder configured")
	}

	data, err := ioutil.ReadAll(e.r)
	if err != nil {
		return err
	}

	vals, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	err = e.dec.Decode(v, vals)
	if err != nil {
		return fmt.Errorf("failed to decode into %v from %v: %v", v, vals, err)
	}

	return err
}

type formEncoderFactory struct {
	enc FormEncodeProvider
}

//MimeType will report the EncodingMimeType
func (e *formEncoderFactory) MimeType() string { return MediaTypeForm }

//Encoder will create encoders
func (e *formEncoderFactory) Encoder(w io.Writer) Encoder { return &FormEncoder{e.enc, w} }

type formDecoderFactory struct {
	dec FormDecodeProvider
}

//MimeType will report the EncodingMimeType
func (e *formDecoderFactory) MimeType() string { return MediaTypeForm }

//Decoder will create decoders
func (e *formDecoderFactory) Decoder(r io.Reader) Decoder { return &FormDecoder{e.dec, r} }

//NewFormEncoding creates the factory using a provider, often third party library
func NewFormEncoding(p FormEncodeProvider) EncoderFactory {
	return &formEncoderFactory{p}
}

//NewFormDecoding creates the factory using a provider, often third party library
func NewFormDecoding(p FormDecodeProvider) DecoderFactory {
	return &formDecoderFactory{p}
}
