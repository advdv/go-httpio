package handling

import (
	"context"
	"net/http"

	"github.com/advanderveer/go-httpio/encoding"
)

//ValidationErr encapulate a validation error
type ValidationErr struct{ error }

//StatusCode allows parse error to return bad request status
func (e ValidationErr) StatusCode() int { return http.StatusBadRequest }

//Cause is the underlying error of the parsing error
func (e ValidationErr) Cause() error { return e.error }

//DecodeErr is returned during decoding and implements the causer interface which can later
//be used to find the cause of the error (custom or otherwise)
type DecodeErr struct{ error }

//StatusCode allows parse error to return bad request status
func (e DecodeErr) StatusCode() int { return http.StatusBadRequest }

//Cause is the underlying error of the parsing error
func (e DecodeErr) Cause() error { return e.error }

var (
	//HeaderHandlingError is set whenever the server knows it has encountered and handling error
	HeaderHandlingError = "X-Has-Handling-Error"

	//HeaderErrHandling will simply return the generic Err struct for encoding that includes a
	//statuscode that defaults to 500 unless the input error specifies otherwise, it uses an header
	//value to signal to the client that the body contains an error.
	HeaderErrHandling = func(ctx context.Context, err error, whdr http.Header) interface{} {
		e := &Err{Message: err.Error()}
		if errs, ok := err.(Statuser); ok {
			e.status = errs.StatusCode()
		} else {
			e.status = http.StatusInternalServerError
		}

		whdr.Set(HeaderHandlingError, "1")
		return e
	}

	//HeaderErrReceiver uses the header value to determine if the response holds an error. It should
	//rturn nil if the response contains no error
	HeaderErrReceiver = func(ctx context.Context, resp *http.Response) error {
		if resp.Header.Get(HeaderHandlingError) != "" {
			errOut := &Err{status: resp.StatusCode} //we will decode into the general Err struct instead
			return errOut
		}

		return nil
	}
)

//ErrHandler provides a central location to gather errors, it gets passed the request context and
//the error that was encountered. It is expected to return a value that can then be encoded through
//normal means and is allowed to change the response heade whdr. If complete control over how the
//error will be rendered is required it is possible for the returned value to implement the
//Renderable interface. As a special case the returned value can implement the Statuser interface
//to customize the status code that will be returned
type ErrHandler func(ctx context.Context, err error, whdr http.Header) interface{}

//ErrReceiver is used by the client to determine if the response holds an error value and specify the
//struct to decode it into. If the reponse holds an error this function should return a non-nil value
type ErrReceiver func(ctx context.Context, resp *http.Response) error

//Err is the struct that is encoded when a generic error
//value reaches the render method
type Err struct {
	status  int
	Message string `json:"message" form:"message" schema:"message"`
}

//Error lets this be used on the client side to return as an actual error
func (e Err) Error() string { return e.Message }

//StatusCode sets the status that is used whenever the error is encoded
func (e Err) StatusCode() int { return e.status }

//Validator provides validation just after decoding
type Validator interface {
	Validate(v interface{}) error
}

//H allows http request parsing and output writing using encoding stacks
type H struct {
	encs       encoding.Stack
	ErrHandler ErrHandler
	Validator  Validator
}

//NewH will setup handling using encoding stack 'encs'
func NewH(encs encoding.Stack) *H {
	return &H{encs, HeaderErrHandling, nil}
}
