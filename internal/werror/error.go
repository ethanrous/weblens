package werror

import (
	"errors"
	"fmt"
)

func Errorf(format string, args ...any) StackError {
	return &withStack{
		err:   fmt.Errorf(format, args...),
		stack: callers(3),
	}
}

var NotImplemented = func(note string) error {
	return &withStack{
		err:   fmt.Errorf("not implemented: %s", note),
		stack: callers(3),
	}
}

// ClientSafeErr packages an error that is safe to send to the client, and the real error that should be logged on the server.
// See TrySafeErr for more information
type ClientSafeErr struct {
	realError  error
	safeErr    error
	arg        any
	statusCode int
}

func (cse ClientSafeErr) Error() string {
	if cse.realError == nil {
		return cse.Safe().Error()
	} else if cse.arg != nil {
		return fmt.Sprintf("%s: %v", cse.realError.Error(), cse.arg)
	}
	return cse.realError.Error()
}

func (cse ClientSafeErr) Code() int {
	return cse.statusCode
}

func (cse ClientSafeErr) Safe() error {
	if cse.safeErr == nil {
		return errors.New("Unknown Server Error")
	}
	return cse.safeErr
}

func (cse ClientSafeErr) Unwrap() error {
	if cse.realError == nil {
		return cse.Safe()
	}
	return cse.realError
}

func (cse ClientSafeErr) WithArg(arg any) ClientSafeErr {
	newCse := cse
	newCse.realError = cse
	newCse.arg = arg
	return newCse
}
