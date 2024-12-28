package common

type Result[T any] struct {
	Value T
	Error error
}

func (r Result[T]) Match() (T, error) {
	return r.Value, r.Error
}
