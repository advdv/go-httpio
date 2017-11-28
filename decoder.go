package httpio

import (
	"io"
)

//Decoder allows for values to be encoded
type Decoder interface {
	Decode(v interface{}) error
}

//DecoderFactory creates encoders for a writer
type DecoderFactory interface {
	MimeType() string
	Decoder(r io.Reader) Decoder
}

//DecoderList offers encoder factories
type DecoderList []DecoderFactory

//Default returns the first avaible encoding set
func (s DecoderList) Default() DecoderFactory {
	if len(s) < 1 {
		panic("No decode factories configured")
	}
	return s[0]
}

//Supported lists all media types by the encoding stack in order of preference
func (s DecoderList) Supported() (supported []string) {
	for _, enc := range s {
		supported = append(supported, enc.MimeType())
	}
	return supported
}

//Find an encoding mechanism by its mime type: O(N). Returns nil if none is found
func (s DecoderList) Find(mime string) DecoderFactory {
	for _, e := range s {
		if e.MimeType() == mime {
			return e
		}
	}

	return nil
}
