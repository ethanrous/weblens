package errors

import (
	"errors"
)

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

func Statusf(code int, format string, args ...any) error {
	err := Errorf(format, args...)

	return &statusError{
		code: code,
		err:  err,
	}
}

func WrapStatus(code int, err error) error {
	return &statusError{
		code: code,
		err:  err,
	}
}
