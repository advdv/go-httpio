package handling

import "github.com/advanderveer/go-httpio/encoding"

var (
	//HeaderHandlingError is set whenever the server knows it has encountered and handling error
	HeaderHandlingError = "X-Has-Handling-Error"
)

//Err is the struct that is encoded when a generic error
//value reaches the render method
type Err struct {
	Message string `json:"message" form:"message" schema:"message"`
}

//H allows http request parsing and output writing using encoding stacks
type H struct {
	encs encoding.Stack
}

//NewH will setup handling using encoding stack 'encs'
func NewH(encs encoding.Stack) *H {
	return &H{encs}
}
