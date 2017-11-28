package middleware

import (
	"encoding/json"
	"io"
)

var (
	//MediaTypeJSON identifies JSON content
	MediaTypeJSON = "application/json"
)

//JSON allows encode and decode into JSOn
type JSON struct{}

//MimeType will report the EncodingMimeType
func (e *JSON) MimeType() string { return MediaTypeJSON }

//Encoder will create encoders
func (e *JSON) Encoder(w io.Writer) Encoder { return json.NewEncoder(w) }

//Decoder will create decoders
func (e *JSON) Decoder(r io.Reader) Decoder { return json.NewDecoder(r) }
