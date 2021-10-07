package logs

type Exception struct {
	Meta interface{}
	Err  error
}

func (e Exception) Error() string {
	return e.Err.Error()
}
func (e Exception) Unwrap() error {
	return e.Err
}
