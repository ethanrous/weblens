package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

type emptyLogger struct{}

func (emptyLogger) Printf(format string, v ...any) {}
func (emptyLogger) Println(v ...any)               {}

// Info writes logs in the color blue with "INFO: " as prefix
var Info = log.New(os.Stdout, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)
var WsInfo = log.New(os.Stdout, "\u001b[34m[WS INFO] \u001B[0m", log.LstdFlags)

// Warning writes logs in the color yellow with "WARNING: " as prefix
var Warning = log.New(os.Stdout, "\u001b[33m[WARNING] \u001B[0m", log.LstdFlags|log.Lshortfile)

// Error writes logs in the color red with "ERROR: " as prefix
var Error = log.New(os.Stdout, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Llongfile)

// ErrorCatcher Same as error, but don't print the file and line. It is expected that what is being printed will include a file and line
// Useful for a generic panic cather function, where the line of the catcher function is not useful
var ErrorCatcher = log.New(os.Stdout, "\u001b[31m[ERROR] \u001B[0m", log.LstdFlags)

var Debug Logger = emptyLogger{}
var Trace Logger = emptyLogger{}

const (
	ERRORS_ONLY = -1
	DEFAULT     = 0
	DEBUG       = 1
	TRACE       = 2
)

var logLevel = 0

func SetLogLevel(level int) {
	if logLevel == 0 {
		logLevel = level
	}
	switch logLevel {
	// Disable all loggers except for error
	case ERRORS_ONLY:
		Info = log.New(io.Discard, "", 0)
		Warning = log.New(io.Discard, "", 0)
	// do nothing
	case DEFAULT:
	// enable trace and debug
	case TRACE:
		Trace = log.New(os.Stdout, "[TRACE] ", log.LstdFlags|log.Lshortfile)
		fallthrough
	// enable debug
	case DEBUG:
		prefix := fmt.Sprintf("\u001b[36m[%s] \u001B[0m", "DEBUG")
		Debug = log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)

	}
}
