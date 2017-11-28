package httpio

import (
	"io"
)

//Encoder allows for values to be encoded
type Encoder interface {
	Encode(v interface{}) error
}

//EncoderFactory creates encoders for a writer
type EncoderFactory interface {
	MimeType() string
	Encoder(w io.Writer) Encoder
}

//EncoderList offers encoder factories
type EncoderList []EncoderFactory

//Default returns the first avaible encoding set
func (s EncoderList) Default() EncoderFactory {
	if len(s) < 1 {
		panic("No encoder factories configured")
	}
	return s[0]
}

//Supported lists all media types by the encoding stack in order of preference
func (s EncoderList) Supported() (supported []string) {
	for _, enc := range s {
		supported = append(supported, enc.MimeType())
	}
	return supported
}

//Find an encoding mechanism by its mime type: O(N). Returns
//nil if none is found
func (s EncoderList) Find(mime string) EncoderFactory {
	for _, e := range s {
		if e.MimeType() == mime {
			return e
		}
	}

	return nil
}
