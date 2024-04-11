package logs

import (
	"github.com/go-errors/errors"
)

type ExceptionError struct {
	meta interface{}
	Err  error
}

func NewException(err error, meta interface{}) ExceptionError {
	var v *errors.Error
	if !errors.As(err, &v) {
		v = errors.Wrap(err, 1)
	}
	return ExceptionError{
		meta: meta,
		Err:  v,
	}
}
func NewExceptionWithMeta(err error, kvList ...interface{}) ExceptionError {

	if len(kvList)%2 != 0 {
		kvList = append(kvList, "<no-value>")
	}
	var subMeta map[string]interface{}
	if vE, ok := err.(ExceptionError); ok {
		subMeta = vE.GetMeta()
		err = vE.Err
	} else {
		var se *errors.Error
		if !errors.As(err, &se) {
			err = errors.Wrap(err, 1)
		}
	}
	meta := make(map[string]interface{}, (len(kvList)/2)+len(subMeta))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			continue
		}
		meta[k] = kvList[i+1]
	}

	for s, i := range subMeta {
		meta[s] = i
	}
	return ExceptionError{
		meta: meta,
		Err:  err,
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
	if e.meta == nil {
		return nil
	}
	return e.meta.(map[string]interface{})
}
