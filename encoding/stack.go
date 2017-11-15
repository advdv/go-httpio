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
		return nil, errors.New("No encoding configured")
	}
	return s[0], nil
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
