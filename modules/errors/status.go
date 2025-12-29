package errors

import (
	"errors"
)

// StatusErr is an error that includes an HTTP status code.
type StatusErr interface {
	error
	Status() int
}

type statusError struct {
	code int
	err  error
}

func (e *statusError) Error() string {
	return e.err.Error()
}

func (e *statusError) Unwrap() error {
	return e.err
}

func (e *statusError) Status() int {
	return e.code
}

// AsStatus extracts the HTTP status code from an error, returning the default status if the error does not implement StatusErr.
func AsStatus(err error, defaultStatus int) (int, string) {
	if err == nil {
		return 0, ""
	}

	var statusErr StatusErr
	if errors.As(err, &statusErr) {
		return statusErr.Status(), statusErr.Error()
	}

	return defaultStatus, err.Error()
}

// Statusf creates a new error with the specified HTTP status code and formatted message.
func Statusf(code int, format string, args ...any) error {
	err := Errorf(format, args...)

	return &statusError{
		code: code,
		err:  err,
	}
}

// WrapStatus wraps an existing error with an HTTP status code.
func WrapStatus(code int, err error) error {
	return &statusError{
		code: code,
		err:  err,
	}
}
