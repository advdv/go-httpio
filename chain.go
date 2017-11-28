package httpio

import "net/http"

//TransFunc implements Transformer when casted to
type TransFunc func(a interface{}, r *http.Request, w http.ResponseWriter) error

//Transform allows a transfunc to be used as a Transformer
func (f TransFunc) Transform(a interface{}, r *http.Request, w http.ResponseWriter) error {
	return f(a, r, w)
}

//Transformer is used to transform value a in the context of responding with 'w' to request 'r'
type Transformer interface {
	Transform(a interface{}, r *http.Request, w http.ResponseWriter) error
}

//Transware is used to implement a chain of transformers, works like router middlewares
type Transware func(next Transformer) Transformer

//Chain builds a recursing transformer with 'base' at the end and 'others' in front. If any transformer
//returns an error the recursion is unwound and the error is returned.
func Chain(base Transformer, others ...Transware) Transformer {
	if len(others) == 0 {
		return base
	}

	return others[0](Chain(base, others[1:cap(others)]...))
}
