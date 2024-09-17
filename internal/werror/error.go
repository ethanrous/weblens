package werror

import (
	"errors"
	"fmt"

	"github.com/ethrousseau/weblens/internal/log"
)

func Errorf(format string, args ...any) StackError {
	return &withStack{
		err:   fmt.Errorf(format, args...),
		stack: callers(),
	}
}

var NotImplemented = func(note string) error {
	return &withStack{
		err:   fmt.Errorf("not implemented: %s", note),
		stack: callers(),
	}
}

type clientSafeErr struct {
	realError  error
	safeErr    error
	statusCode int
}

func (cse *clientSafeErr) Error() string {
	if cse.realError == nil {
		return cse.Safe().Error()
	}
	return cse.realError.Error()
}

func (cse *clientSafeErr) Safe() error {
	if cse.safeErr == nil {
		return errors.New("Unknown Server Error")
	}
	return cse.safeErr
}

func TrySafeErr(err error) (error, int) {
	if err == nil {
		return nil, 200
	}

	var safeErr = &clientSafeErr{}
	if errors.As(err, &safeErr) {
		log.ShowErr(err)
		return safeErr.Safe(), safeErr.statusCode
	}

	log.ErrTrace(err)
	return errors.New("Unknown Server Error"), 500
}
