package logs

import (
	"github.com/go-errors/errors"
	"reflect"
)

type Exception struct {
	meta interface{}
	Err  error
}

func NewException(err error, meta interface{}) Exception {
	v, ok := err.(*errors.Error)
	if !ok {
		v = errors.Wrap(err, 1)
	}
	return Exception{
		meta: meta,
		Err:  v,
	}

}
func (e Exception) Error() string {
	return e.Err.Error()
}
func (e Exception) Unwrap() error {
	return e.Err
}
func (e Exception) ErrorStack() string {
	v, ok := e.Err.(*errors.Error)
	if !ok {
		v = errors.Wrap(e.Err, 2)
	}
	return v.ErrorStack()
}

func (e Exception) GetMeta() map[string]interface{} {
	var maps = make(map[string]interface{})
	if e.meta == nil {
		return nil
	}
	maps["level_1"] = map[string]interface{}{reflect.TypeOf(e.meta).String(): e.meta}
	if v, ok := e.Err.(Exception); ok {
		meta := v.GetMeta()
		if meta != nil {
			maps["level_2"] = meta
		}
	}
	return maps
}
