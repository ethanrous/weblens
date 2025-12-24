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

func (o Option[T]) Get() (value T, isSet bool) {
	return o.value, o.present
}

func (o Option[T]) GetOr(fallback T) (value T) {
	if o.present {
		return o.value
	}

	return fallback
}

func (o Option[T]) Set(value T) Option[T] {
	o.value = value
	o.present = true

	return o
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.present {
		return json.Marshal(o.value)
	}

	return json.Marshal(nil)
}
