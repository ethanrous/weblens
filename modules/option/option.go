// Package option provides a generic Option type for representing optional values.
package option

import (
	"encoding/json"
)

// Option represents an optional value that may or may not be present.
type Option[T any] struct {
	value   T
	present bool
}

// Of creates an Option containing the given value.
func Of[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

// Zero creates an empty Option with no value.
func Zero[T any]() Option[T] {
	return Option[T]{present: false}
}

// Has returns true if the Option contains a value.
func (o Option[T]) Has() bool {
	return o.present
}

// Get returns the value and whether it is present.
func (o Option[T]) Get() (value T, isSet bool) {
	return o.value, o.present
}

// GetOr returns the value if present, otherwise returns the fallback.
func (o Option[T]) GetOr(fallback T) (value T) {
	if o.present {
		return o.value
	}

	return fallback
}

// Set sets the value and returns the updated Option.
func (o Option[T]) Set(value T) Option[T] {
	o.value = value
	o.present = true

	return o
}

// MarshalJSON implements the json.Marshaler interface.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.present {
		return json.Marshal(o.value)
	}

	return json.Marshal(nil)
}
