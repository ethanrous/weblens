package werror

import (
	"fmt"

	"github.com/pkg/errors"
)

func Errorf(format string, args ...any) error {
	return errors.New(fmt.Sprintf(format, args...))
}

var NotImplemented = func(note string) error {
	return &withStack{
		err:   fmt.Errorf("not implemented: %s", note),
		stack: callers(3),
	}
}

func NewClientSafeError(safeErr error, statusCode int, realErr ...error) ClientSafeErr {
	var rerr error
	if len(realErr) > 0 {
		rerr = realErr[0]
	}

	return ClientSafeErr{
		safeErr:    safeErr,
		statusCode: statusCode,
		realError:  rerr,
	}
}

// ClientSafeErr packages an error that is safe to send to the client, and the real error that should be logged on the server.
// See TrySafeErr for more information
type ClientSafeErr struct {
	realError  error
	safeErr    error
	arg        any
	skipTrace  bool
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

func (cse ClientSafeErr) WithArg(arg ...any) ClientSafeErr {
	newCse := cse
	newCse.realError = cse
	newCse.arg = arg
	return newCse
}

func GetSafeErr(err error) (error, int) {
	var safeErr = ClientSafeErr{}
	if errors.As(err, &safeErr) {
		return safeErr.Safe(), safeErr.Code()
	}

	return errors.New("Unknown Server Error"), 500
}
