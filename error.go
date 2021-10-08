package logs

import (
	"reflect"
)

type Exception struct {
	meta interface{}
	Err  error
}

func NewException(err error, meta interface{}) Exception {
	return Exception{
		meta: meta,
		Err:  err,
	}

}
func (e Exception) Error() string {
	return e.Err.Error()
}
func (e Exception) Unwrap() error {
	return e.Err
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
