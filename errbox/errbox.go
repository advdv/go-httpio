package errbox

import "errors"

//ErrBoxer allows storing an error
type ErrBoxer interface {
	BoxError(err error)
}

//Err can be embedded into output structs to enable servers to
//send error messages over to the client as a string:
type Err string

//BoxError implements a method that allows any type that embeds E
//to allow basic piggybacking of error values
func (err *Err) BoxError(e error) {
	*err = Err(e.Error())
}

//UnboxErr returns nil if the embedding struct doesn't have an error value
//embedded or the actual error if it does
func (err *Err) UnboxErr() error {
	if err == nil || *err == "" {
		return nil
	}

	return errors.New(string(*err))
}

//Box will assert if value 'v' allows error 'err' to be stored inside it. If 'err' is nil it
//returns 'v'. If 'v' can box 'err' it will return 'v' else it returns 'err'
func Box(v interface{}, err error) (interface{}, bool) {
	if err == nil {
		return v, true
	}

	if errBoxer, ok := v.(ErrBoxer); ok {
		errBoxer.BoxError(err)
		return errBoxer, true
	}

	return err, false
}
