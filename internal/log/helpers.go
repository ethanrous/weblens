package log

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type StackError interface {
	Error() string
	Stack() string
	Errorln() string
}

func ErrTrace(err error, extras ...string) {
	if err != nil {
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

		msg := strings.Join(extras, " ")
		_, file, line, _ := runtime.Caller(1)
		ErrorCatcher.Printf("%s:%d %s: %s", file, line, msg, err)
	}
}

func colorStatus(status int) string {
	if status < 400 {
		return fmt.Sprintf("\u001b[32m%d\u001B[0m", status)
	} else if status >= 400 && status < 500 {
		return fmt.Sprintf("\u001b[33m%d\u001B[0m", status)
	} else if status >= 500 {
		return fmt.Sprintf("\u001b[31m%d\u001B[0m", status)
	}
	return "Not reached"
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

func ApiLogger(logLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.RequestURI

		handler := runtime.FuncForPC(reflect.ValueOf(c.Handler()).Pointer()).Name()
		handler = handler[strings.LastIndex(handler, ".")+1:]

		c.Next()

		status := c.Writer.Status()
		if logLevel == -1 && status < 400 {
			return
		}

		remote := c.ClientIP()
		method := c.Request.Method
		timeTotal := time.Since(start)

		metrics.RequestsTimer.With(
			prometheus.Labels{
				"handler": handler, "method": c.Request.Method,
			},
		).Observe(timeTotal.Seconds())

		fmt.Printf(
			"\u001B[0m[%s][%s][%7s][%s][%s] %s %s\n", start.Format("Jan 02 15:04:05"), remote,
			colorTime(timeTotal),
			handler, colorStatus(status), method, path,
		)
	}
}
