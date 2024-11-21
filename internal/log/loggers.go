package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger
var ErrorCatcher Logger

func init() {
	Info = log.New(os.Stdout, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)

	// Warning writes logs in the color yellow with "WARN: " as prefix
	Warning = log.New(os.Stdout, "\u001b[33m[WARN] \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(os.Stdout, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Llongfile)

	// ErrorCatcher Same as error, but don't print the file and line. It is expected that what is being printed will include a file and line
	// Useful for a generic panic cather function, where the line of the catcher function is not useful
	// ErrorCatcher = log.New(os.Stdout, "\u001b[31m[ERROR] \u001B[0m", log.LstdFlags)
	ErrorCatcher = &logger{prefix: "\u001b[31m[ERROR] \u001B[0m", defaultSkip: 4}
}

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05") + " "
}

type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
	Printfn(skip int, format string, v ...any)
	Printlnn(skip int, v ...any)
}

type FuncLogger interface {
	Logger

	// Func allows for a logger callback, that can be optionally disabled based on the log level.
	// This way, expensive logging operations can be avoided if the log level is too low by putting them
	// inside the function.
	Func(func(Logger))
}

type logger struct {
	prefix      string
	defaultSkip int
}

func (l *logger) Printf(format string, v ...any) {
	l.Printfn(l.defaultSkip, format, v...)
}

func (l *logger) Println(v ...any) {
	l.Printlnn(l.defaultSkip, v...)
}

func (l *logger) Printfn(skip int, format string, v ...any) {
	fmt.Println(l.prefix + formatTime() + fmtCaller(skip) + ": " + fmt.Sprintf(format, v...))
}

func (l *logger) Printlnn(skip int, v ...any) {
	fmt.Print(l.prefix + formatTime() + fmtCaller(skip) + ": " + fmt.Sprintln(v...))
}

func NewLogger(prefix string, skip int) Logger {
	return &logger{prefix: prefix, defaultSkip: skip}
}

type funcLogger struct {
	Logger
}

func (f *funcLogger) Printf(format string, v ...any) {
	f.Logger.Printfn(3, format, v...)
}

func (f *funcLogger) Println(v ...any) {
	f.Logger.Printlnn(3, v...)
}

func (f *funcLogger) Func(fn func(l Logger)) {
	fn(f.Logger)
}

type emptyLogger struct{}

func (emptyLogger) Printf(format string, v ...any)            {}
func (emptyLogger) Println(v ...any)                          {}
func (emptyLogger) Printfn(skip int, format string, v ...any) {}
func (emptyLogger) Printlnn(skip int, v ...any)               {}
func (emptyLogger) Func(fn func(l Logger))                    {}

var Debug FuncLogger = emptyLogger{}
var Trace FuncLogger = emptyLogger{}

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
		Trace = &funcLogger{Logger: NewLogger("[TRACE] ", 3)}
		Trace.Println("Trace logger enabled")
		fallthrough
	// enable debug
	case DEBUG:
		prefix := fmt.Sprintf("\u001b[36m[%s] \u001B[0m", "DEBUG")
		// Debug = log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)
		Debug = &funcLogger{Logger: NewLogger(prefix, 3)}

	}

	Debug.Printf("Using log level [%d]", newLevel)
}

func fmtCaller(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	file = file[strings.LastIndex(file, "/")+1:]
	return file + ":" + strconv.Itoa(line)
}

func TraceCaller(skip int, format string, v ...any) {
	if logLevel >= TRACE {
		_, file, line, _ := runtime.Caller(2)

		file = file[strings.LastIndex(file, "/")+1:]

		fmt.Printf("[TRACE] %s:%d: %s\n", file, line, fmt.Sprintf(format, v...))
	}

}
