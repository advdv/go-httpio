package httpio

import (
	"encoding/xml"
	"io"
)

var (
	//MediaTypeXML identifies XML content
	MediaTypeXML = "application/xml"
)

//XML allows encode and decode into JSOn
type XML struct{}

//MimeType will report the EncodingMimeType
func (e *XML) MimeType() string { return MediaTypeXML }

//Encoder will create encoders
func (e *XML) Encoder(w io.Writer) Encoder { return xml.NewEncoder(w) }

//Decoder will create decoders
func (e *XML) Decoder(r io.Reader) Decoder { return xml.NewDecoder(r) }
