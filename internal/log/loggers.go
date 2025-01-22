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

	"github.com/go-chi/chi/v5/middleware"
)

var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger
var ErrorCatcher *log.Logger

var Debug FuncLogger = emptyLogger{}
var Trace FuncLogger = emptyLogger{}

func init() {
	Info = log.New(os.Stdout, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)

	// Warning writes logs in the color yellow with "WARN: " as prefix
	Warning = log.New(os.Stdout, "\u001b[33m[WARN] \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(os.Stdout, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Lshortfile)

	// ErrorCatcher Same as error, but don't print the file and line. It is expected that what is being printed will include a file and line
	// Useful for a generic panic cather function, where the line of the catcher function is not useful
	// ErrorCatcher = log.New(os.Stdout, "\u001b[31m[ERROR] \u001B[0m", log.LstdFlags)
	ErrorCatcher = log.New(os.Stdout, "", 0)

	// ErrorCatcher = &logger{prefix: "\u001b[31m[ERROR] \u001B[0m", defaultSkip: 4}
}

// var ErrorCatcher Logger
var Output *os.File

type Bundle struct {
	Trace   FuncLogger
	Debug   FuncLogger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Raw     *log.Logger

	Level Level
}

func (lp Bundle) ErrTrace(err error, extras ...string) {
	if err != nil {
		if lp.Level < DEBUG {
			lp.ShowErr(err, extras...)
			return
		}

		fmter, ok := err.(StackError)
		if ok {
			lp.Raw.Println(fmter.Stack())
			return
		}

		middleware.PrintPrettyStack(err)
	}
}

func (lp Bundle) ShowErr(err error, extras ...string) {
	if err != nil {
		fmter, ok := err.(StackError)
		if ok {
			errStr := fmter.Errorln()
			if errStr[len(errStr)-1] == '\n' {
				errStr = errStr[:len(errStr)-1]
			}
			lp.Raw.Println(errStr)
			return
		}

		msg := ""
		if len(extras) > 0 {
			msg = " " + strings.Join(extras, " ")
		}

		_, file, line, _ := runtime.Caller(2)
		file = file[strings.LastIndex(file, "/")+1:]

		lp.Raw.Printf("%s:%d%s: %s", file, line, msg, err)
	}
}

func NewLogPackage(outputPath string, level Level) Bundle {
	var output *os.File
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			panic(err)
		}
		output = f

	} else {
		output = os.Stdout
		outputPath = "STDOUT"
	}

	fmt.Printf("New logger to [%s] with level %s\n", outputPath, level)

	info := log.New(output, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)
	warning := log.New(output, "\u001b[33m[WARN] \u001B[0m", log.LstdFlags|log.Lshortfile)
	error := log.New(output, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Llongfile)
	errorCatcher := log.New(output, "", 0)

	var trace FuncLogger = emptyLogger{}
	var debug FuncLogger = emptyLogger{}

	if level == TRACE {
		trace = &funcLogger{Logger: NewLogger("[TRACE] ", 3, output)}
		trace.Println("Trace logger enabled")
	}
	if level >= DEBUG {
		prefix := fmt.Sprintf("\u001b[36m[%s] \u001B[0m", "DEBUG")
		debug = &funcLogger{Logger: NewLogger(prefix, 3, output)}
	}
	return Bundle{
		Trace:   trace,
		Debug:   debug,
		Info:    info,
		Warning: warning,
		Error:   error,
		Raw:     errorCatcher,

		Level: level,
	}
}

func NewEmptyLogPackage() Bundle {
	return Bundle{
		Trace:   emptyLogger{},
		Debug:   emptyLogger{},
		Info:    log.New(io.Discard, "", 0),
		Warning: log.New(io.Discard, "", 0),
		Error:   log.New(io.Discard, "", 0),
		Raw:     log.New(io.Discard, "", 0),
	}
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
	output      *os.File
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
	fmt.Fprintln(Output, l.prefix+formatTime()+fmtCaller(skip)+": "+fmt.Sprintf(format, v...))
}

func (l *logger) Printlnn(skip int, v ...any) {
	fmt.Fprint(Output, l.prefix+formatTime()+fmtCaller(skip)+": "+fmt.Sprintln(v...))
}

func NewLogger(prefix string, skip int, output *os.File) Logger {
	return &logger{prefix: prefix, defaultSkip: skip, output: output}
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

type Level string

const (
	DEFAULT Level = "default"
	DEBUG   Level = "debug"
	TRACE   Level = "trace"
)

var logLevel Level = DEFAULT

func GetLogLevel() Level {
	return logLevel
}

func SetLogLevel(newLevel Level, outputPath string) {
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			panic(err)
		}
		Output = f

		Info = log.New(Output, "\u001b[34m[INFO] \u001B[0m", log.LstdFlags)
		Warning = log.New(Output, "\u001b[33m[WARN] \u001B[0m", log.LstdFlags|log.Lshortfile)
		Error = log.New(Output, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Llongfile)
	} else {
		Output = os.Stdout
	}

	if logLevel == newLevel {
		return
	}
	if logLevel != DEFAULT {
		Warning.Println("Overwriting Log level")
	}
	logLevel = newLevel
	Info.Println("Setting log level to", newLevel)

	switch logLevel {
	// do nothing
	case DEFAULT:
	// enable trace and debug
	case TRACE:
		Trace = &funcLogger{Logger: NewLogger("[TRACE] ", 3, Output)}
		Trace.Println("Trace logger enabled")
		fallthrough
	// enable debug
	case DEBUG:
		prefix := fmt.Sprintf("\u001b[36m[%s] \u001B[0m", "DEBUG")
		// Debug = log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)
		Debug = &funcLogger{Logger: NewLogger(prefix, 3, Output)}

	}
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
