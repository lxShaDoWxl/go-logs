package logs

import (
	"reflect"

	"github.com/go-errors/errors"
)

type ExceptionError struct {
	meta interface{}
	Err  error
}

func NewException(err error, meta interface{}) ExceptionError {
	v, ok := err.(*errors.Error)
	if !ok {
		v = errors.Wrap(err, 1)
	}
	return ExceptionError{
		meta: meta,
		Err:  v,
	}
}
func (e ExceptionError) Error() string {
	return e.Err.Error()
}
func (e ExceptionError) Unwrap() error {
	return e.Err
}
func (e ExceptionError) ErrorStack() string {
	v, ok := e.Err.(*errors.Error)
	if !ok {
		v = errors.Wrap(e.Err, 2)
	}
	return v.ErrorStack()
}

func (e ExceptionError) GetMeta() map[string]interface{} {
	var maps = make(map[string]interface{})
	if e.meta == nil {
		return nil
	}
	maps["level_1"] = map[string]interface{}{reflect.TypeOf(e.meta).String(): e.meta}
	if v, ok := e.Err.(ExceptionError); ok {
		meta := v.GetMeta()
		if meta != nil {
			maps["level_2"] = meta
		}
	}
	return maps
}
