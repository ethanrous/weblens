package option

import (
	"encoding/json"
)

type Option[T any] struct {
	value   T
	present bool
}

func Of[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

func Zero[T any]() Option[T] {
	return Option[T]{present: false}
}

func (o Option[T]) Has() bool {
	return o.present
}

func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.present {
		return json.Marshal(o.value)
	}
	return json.Marshal(nil)
}
