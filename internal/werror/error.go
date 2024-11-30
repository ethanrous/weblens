package werror

import (
	"errors"
	"fmt"

	"github.com/ethanrous/weblens/internal/log"
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

// clientSafeErr packages an error that is safe to send to the client, and the real error that should be logged on the server.
// See TrySafeErr for more information
type clientSafeErr struct {
	realError  error
	safeErr    error
	statusCode int
	arg        any
}

func (cse clientSafeErr) Error() string {
	if cse.realError == nil {
		return cse.Safe().Error()
	} else if cse.arg != nil {
		return fmt.Sprintf("%s: %v", cse.realError.Error(), cse.arg)
	}
	return cse.realError.Error()
}

func (cse clientSafeErr) Safe() error {
	if cse.safeErr == nil {
		return errors.New("Unknown Server Error")
	}
	return cse.safeErr
}

func (cse clientSafeErr) Unwrap() error {
	if cse.realError == nil {
		return cse.Safe()
	}
	return cse.realError
}

func (cse clientSafeErr) WithArg(arg any) clientSafeErr {
	newCse := cse
	newCse.realError = cse
	newCse.arg = arg
	return newCse
}

// SafeErr unpackages an error, if possible, to find the error inside that is safe to send to the client.
// If the error is not a clientSafeErr, it will trace the original error in the server logs, and return a generic error
// and a 500 to the client
// The reasoning behind this is, for example, if a user tries to access a file that they aren't allowed to, WE want to know (and log)
// they were not allowed to. Then, we want to tell the client (lie) that the file doesn't exist. This way, we don't give the forbidden
// user any information about the file.
func TrySafeErr(err error) (error, int) {
	if err == nil {
		return nil, 200
	}

	var safeErr = clientSafeErr{}
	if errors.As(err, &safeErr) {
		if safeErr.statusCode >= 400 {
			if log.GetLogLevel() == log.TRACE {
				log.ErrTrace(err)
			} else {
				log.ShowErr(err)
			}
		}
		return safeErr.Safe(), safeErr.statusCode
	}

	log.ErrTrace(err)
	return errors.New("Unknown Server Error"), 500
}
