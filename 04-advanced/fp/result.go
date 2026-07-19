package fp

type Result[E any] struct {
	value E
	err   error
}

func NewValue[E any](value E) Result[E] {
	return Result[E]{value: value}
}

func NewError[E any](err error) Result[E] {
	return Result[E]{err: err}
}

func (r Result[E]) Get() E {
	return r.value
}

func (r Result[E]) IsError() bool {
	return r.err != nil
}

func (r Result[E]) IsOk() bool {
	return r.err == nil
}

func (r Result[E]) Error() error {
	return r.err
}
