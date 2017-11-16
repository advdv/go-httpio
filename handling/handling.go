package handling

import "github.com/advanderveer/go-httpio/encoding"

//H allows http request parsing and output writing using encoding stacks
type H struct {
	encs encoding.Stack
}

//NewH will setup handling using encoding stack 'encs'
func NewH(encs encoding.Stack) *H {
	return &H{encs}
}
