package log

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type StackError interface {
	Error() string
	Stack() string
	Errorln() string
}

func ErrTrace(err error, extras ...string) {
	if err != nil {
		if logLevel < DEBUG {
			ShowErr(err, extras...)
			return
		}

		fmter, ok := err.(StackError)
		if ok {
			ErrorCatcher.Println(fmter.Stack())
			return
		}

		_, file, no, _ := runtime.Caller(1)
		ErrorCatcher.Println(string(debug.Stack()))
		ErrorCatcher.Printf("%s:%d (no stack) %s", file, no, err.Error())
	}
}

func ShowErr(err error, extras ...string) {
	if err != nil {
		fmter, ok := err.(StackError)
		if ok {
			errStr := fmter.Errorln()
			if errStr[len(errStr)-1] == '\n' {
				errStr = errStr[:len(errStr)-1]
			}
			ErrorCatcher.Println(errStr)
			return
		}

		msg := ""
		if len(extras) > 0 {
			msg = " " + strings.Join(extras, " ")
		}

		_, file, line, _ := runtime.Caller(2)
		file = file[strings.LastIndex(file, "/")+1:]

		ErrorCatcher.Printf("%s:%d%s: %s", file, line, msg, err)
	}
}

func colorStatus(status int) string {
	if status == 0 {
		return fmt.Sprintf("\u001b[31m%d\u001B[0m", status)
	} else if status < 400 {
		return fmt.Sprintf("\u001b[32m%d\u001B[0m", status)
	} else if status >= 400 && status < 500 {
		return fmt.Sprintf("\u001b[33m%d\u001B[0m", status)
	} else if status >= 500 {
		return fmt.Sprintf("\u001b[31m%d\u001B[0m", status)
	}
	return fmt.Sprintf("\u001b[31m%s\u001B[0m", "BAD STATUS CODE")
}

func colorTime(dur time.Duration) string {
	durString := dur.String()
	lastDigitIndex := strings.LastIndexFunc(
		durString, func(r rune) bool {
			return r < 58
		},
	)
	if len(durString) > 7 {
		durString = (durString[:lastDigitIndex+1])[:4] + durString[lastDigitIndex:]
	}

	if dur < time.Millisecond*200 {
		return durString
	} else if dur < time.Second {
		return fmt.Sprintf("\u001b[33m%s\u001B[0m", durString)
	} else {
		return fmt.Sprintf("\u001b[31m%s\u001B[0m", durString)
	}
}

func ApiLogger(logger LogPackage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status >= 400 && status < 500 && ww.BytesWritten() == 0 {
				logger.Error.Println("4xx DID NOT SEND ERROR")
			}

			if status == 0 && r.Header.Get("Upgrade") == "websocket" {
				status = 101
			}

			remote := r.RemoteAddr
			method := r.Method
			timeTotal := time.Since(start)

			route := chi.RouteContext(r.Context()).RoutePattern()

			logger.Info.Printf("\u001B[0m[%s][%7s][%s %s][%s]\n", remote, colorTime(timeTotal), method, route, colorStatus(status))

		})
	}
}
