package encoding

import "errors"

//Stack provides encoding preference, the first encoding is used by default
type Stack []Encoding

//NewStack sets up a stack with a default encoding mechanism
func NewStack(def Encoding, others ...Encoding) Stack {
	s := []Encoding{def}
	s = append(s, others...)
	return s
}

//Default returns the first avaible encoding set
func (s Stack) Default() (Encoding, error) {
	if len(s) < 1 {
		return nil, errors.New("No encodings configured")
	}
	return s[0], nil
}

//Supported lists all media types by the encoding stack in order of preference
func (s Stack) Supported() (supported []string) {
	for _, enc := range s {
		supported = append(supported, enc.MimeType())
	}
	return supported
}

//Find an encoding mechanism by its mime type: O(N). Returns
//nil if none is found
func (s Stack) Find(mime string) Encoding {
	for _, e := range s {
		if e.MimeType() == mime {
			return e
		}
	}

	return nil
}
