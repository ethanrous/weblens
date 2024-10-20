package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger
var ErrorCatcher *log.Logger

func init() {
	Info = log.New(os.Stdout, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)

	// Warning writes logs in the color yellow with "WARN: " as prefix
	Warning = log.New(os.Stdout, "\u001b[33m[WARN] \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(os.Stdout, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Llongfile)

	// ErrorCatcher Same as error, but don't print the file and line. It is expected that what is being printed will include a file and line
	// Useful for a generic panic cather function, where the line of the catcher function is not useful
	ErrorCatcher = log.New(os.Stdout, "\u001b[31m[ERROR] \u001B[0m", log.LstdFlags)
}

type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

type emptyLogger struct{}

func (emptyLogger) Printf(format string, v ...any) {}
func (emptyLogger) Println(v ...any)               {}

var Debug Logger = emptyLogger{}
var Trace Logger = emptyLogger{}

const (
	QUIET   = -1
	DEFAULT = 0
	DEBUG   = 1
	TRACE   = 2
)

var logLevel = 0

func GetLogLevel() int {
	return logLevel
}

func SetLogLevel(newLevel int) {
	if logLevel == newLevel {
		return
	}
	if logLevel != 0 {
		Warning.Println("Overwriting Log level")
	}
	logLevel = newLevel

	switch logLevel {
	// Disable all loggers except for error
	case QUIET:
		Info = log.New(io.Discard, "", 0)
		Warning = log.New(io.Discard, "", 0)
	// do nothing
	case DEFAULT:
	// enable trace and debug
	case TRACE:
		Trace = log.New(os.Stdout, "[TRACE] ", log.LstdFlags|log.Lshortfile)
		Trace.Println("Trace logger enabled")
		fallthrough
	// enable debug
	case DEBUG:
		prefix := fmt.Sprintf("\u001b[36m[%s] \u001B[0m", "DEBUG")
		Debug = log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)

	}

	Debug.Printf("Using log level [%d]", newLevel)
}

func TraceCaller(skip int, format string, v ...any) {
	if logLevel >= TRACE {
		_, file, line, _ := runtime.Caller(2)

		file = file[strings.LastIndex(file, "/")+1:]

		fmt.Printf("[TRACE] %s:%d: %s\n", file, line, fmt.Sprintf(format, v...))
	}

}
