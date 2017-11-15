package encoding

import (
	"encoding/json"
	"encoding/xml"
	"io"
)

//Decoder is used for decoding into v from reader inside the decoder
type Decoder interface {
	Decode(v interface{}) error
}

//Encoder is used to encode v into the writer held in the encoder
type Encoder interface {
	Encode(v interface{}) error
}

//Encoding provides encoders for a certain content type
type Encoding interface {
	MimeType() string
	Encoder(w io.Writer) Encoder
	Decoder(r io.Reader) Decoder
}

//XML allows encode and decode into XML
type XML struct{}

//MimeType will report the EncodingMimeType
func (e *XML) MimeType() string { return "application/xml" }

//Encoder will create encoders
func (e *XML) Encoder(w io.Writer) Encoder { return xml.NewEncoder(w) }

//Decoder will create decoders
func (e *XML) Decoder(r io.Reader) Decoder { return xml.NewDecoder(r) }

//JSON allows encode and decode into JSOn
type JSON struct{}

//MimeType will report the EncodingMimeType
func (e *JSON) MimeType() string { return "application/json" }

//Encoder will create encoders
func (e *JSON) Encoder(w io.Writer) Encoder { return json.NewEncoder(w) }

//Decoder will create decoders
func (e *JSON) Decoder(r io.Reader) Decoder { return json.NewDecoder(r) }
