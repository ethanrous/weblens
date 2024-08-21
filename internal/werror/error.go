package werror

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

type WErr interface {
	Error() string
	ErrorTrace() string
}

type WeblensError struct {
	err        error
	sourceFile string
	sourceLine int
	trace      string
}

func Wrap(err error) WErr {
	wlErr, ok := err.(WErr)
	if !ok {
		return NewWeblensError(err.Error())
	}
	return wlErr
}

func WErrMsg(errMsg string) WErr {
	return NewWeblensError(errMsg)
}

func NewWeblensError(err string) WErr {
	_, filename, line, _ := runtime.Caller(2)
	buf := make([]byte, 1<<16)

	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	return WeblensError{errors.New(err), filepath.Base(filename), line, string(buf)}
}

func (e WeblensError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.sourceFile, e.sourceLine, e.err)
}

func (e WeblensError) ErrorTrace() string {
	return fmt.Sprintf("%s:%d: %s\n%s", e.sourceFile, e.sourceLine, e.err, e.trace)
}

func (e WeblensError) GetSourceFile() string {
	return e.sourceFile
}

func (e WeblensError) GetSourceLine() int {
	return e.sourceLine
}

var NotImplemented = func(note string) WErr {
	return NewWeblensError(
		fmt.Sprint(
			"not implemented: ", note,
		),
	)
}
