package encoding

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
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

//Form allows for encoding and decoding from structs using a third party
//provider such as github.com/gorilla/schema
type Form struct {
	dec FormDecodeProvider
	enc FormEncodeProvider
}

//FormEncoder uses the form encoding provider to implement the Encoding interface
type FormEncoder struct {
	enc FormEncodeProvider
	w   io.Writer
}

//Encode the value v into the encoder writer
func (e *FormEncoder) Encode(v interface{}) error {
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

//MimeType will report the EncodingMimeType
func (e *Form) MimeType() string { return MediaTypeForm }

//Encoder will create encoders
func (e *Form) Encoder(w io.Writer) Encoder { return &FormEncoder{e.enc, w} }

//Decoder will create decoders
func (e *Form) Decoder(r io.Reader) Decoder { return &FormDecoder{e.dec, r} }

//NewFormEncoding will setup encoding and decoding of forms
func NewFormEncoding(enc FormEncodeProvider, dec FormDecodeProvider) *Form {
	return &Form{dec, enc}
}
