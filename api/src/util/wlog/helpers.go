package wlog

import (
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
)

func ErrTrace(err error, extras ...string) {
	if err != nil {
		msg := strings.Join(extras, " ")
		_, file, line, _ := runtime.Caller(1)
		if wlErr, ok := err.(types.WeblensError); ok {
			ErrorCatcher.Printf("(%s:%d) %s", filepath.Base(file), line, wlErr.ErrorTrace())
		} else {
			ErrorCatcher.Printf(
				"%s:%d %s: %s\n----- STACK FOR ERROR ABOVE -----\n%s", file, line, msg, err, debug.Stack(),
			)
		}
	}
}

func ShowErr(err error, extras ...string) {
	if err != nil {
		msg := strings.Join(extras, " ")
		_, file, line, _ := runtime.Caller(1)
		if wlErr, ok := err.(types.WeblensError); ok {
			ErrorCatcher.Printf("(%s:%d) %s", filepath.Base(file), line, wlErr.Error())
		} else {
			ErrorCatcher.Printf("%s:%d %s: %s", file, line, msg, err)
		}
	}
}
