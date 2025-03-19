package werror

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type frame uintptr
type stack []uintptr

func (f frame) pc() uintptr { return uintptr(f) - 1 }

func (f frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

func (f frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func (f frame) String() (str string) {
	return f.name() + "\n\t" + f.file() + ":" + strconv.Itoa(f.line())
}

var (
	red    = "\u001b[31m"
	blue   = "\u001b[32m"
	yellow = "\u001b[33m"
	orange = "\u001b[36m"
	reset  = "\u001B[0m"
)

func formatFramePair(frameStr string) string {
	splitIndex := strings.Index(frameStr, "\n")
	if splitIndex == -1 {
		return ""
	}
	packAndFunc := frameStr[:splitIndex]
	fileAndLine := frameStr[splitIndex:]

	slashIndex := strings.LastIndex(packAndFunc, "/")
	if slashIndex != -1 {
		dotIndex := strings.Index(packAndFunc[slashIndex:], ".")
		packAndFunc = yellow + packAndFunc[:slashIndex+dotIndex] + blue + packAndFunc[slashIndex+dotIndex:] + reset
	} else if strings.HasPrefix(packAndFunc, "panic") {
		packAndFunc = red + packAndFunc + reset
	}

	startIndex := strings.Index(fileAndLine, "/")
	fileIndex := strings.LastIndex(fileAndLine, "/")
	lineIndex := strings.LastIndex(fileAndLine, ":")
	if lineIndex == -1 {
		return "[MALFORMED STACK FRAME] " + frameStr
	}

	fileAndLine = fileAndLine[startIndex:fileIndex+1] + orange + fileAndLine[fileIndex+1:lineIndex] + blue + fileAndLine[lineIndex:] + reset + "\n\n"

	return "    " + packAndFunc + "\n      " + fileAndLine
}

func formatTopFramePair(frameStr string) string {
	splitIndex := strings.Index(frameStr, "\n")
	packAndFunc := frameStr[:splitIndex]
	fileAndLine := frameStr[splitIndex:]

	slashIndex := strings.LastIndex(packAndFunc, "/")
	if slashIndex != -1 {
		dotIndex := strings.Index(packAndFunc[slashIndex:], ".")
		packAndFunc = yellow + packAndFunc[:slashIndex+dotIndex] + red + packAndFunc[slashIndex+dotIndex:] + reset
	}

	startIndex := strings.Index(fileAndLine, "/")
	fileIndex := strings.LastIndex(fileAndLine, "/")
	lineIndex := strings.LastIndex(fileAndLine, ":")
	fileAndLine = fileAndLine[startIndex:fileIndex+1] + red + fileAndLine[fileIndex+1:lineIndex] + blue + fileAndLine[lineIndex:] + reset + "\n"

	return red + "->  " + reset + packAndFunc + "\n" + red + "->    " + reset + fileAndLine + "\n"
}

func (s *stack) String() (stack string) {
	for i, pc := range *s {
		frameStr := frame(pc).String()

		if i == 0 {
			stack += formatTopFramePair(frameStr)
		} else {
			stack += formatFramePair(frameStr)
		}
	}
	return
}

func StackString() string {
	rawStackStr := string(debug.Stack())
	stackStr := ""

	var frame string

	for len(rawStackStr) != 0 {
		firstN := strings.Index(rawStackStr, "\n") + 1
		if firstN == 0 {
			break
		}
		if strings.HasPrefix(rawStackStr, "goroutine") {
			rawStackStr = rawStackStr[firstN:]
			continue
		}

		secondN := strings.Index(rawStackStr[firstN:], "\n") + 1
		if secondN == 0 || secondN == 1 {
			stackStr += "... failed to finish trace ..."
			break
		}
		frame, rawStackStr = rawStackStr[:firstN+secondN], rawStackStr[firstN+secondN:]
		if len(frame) == 0 {
			break
		}

		frame = strings.Trim(frame, "\n")
		stackStr += formatFramePair(frame)
	}

	return stackStr
}

func callers(ignore int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(ignore, pcs[:])

	var st stack = pcs[0:n]
	return &st
}

func GetStack(ignoreDepth int) *stack {
	return callers(ignoreDepth + 3)
}

type StackError interface {
	Error() string
	Stack() string
	Errorln() string
}

var _ StackError = (*withStack)(nil)

type withStack struct {
	err   error
	stack *stack
}

func WithStack(err error) error {
	return errors.WithStack(err)
	// if err == nil {
	// 	return nil
	// }
	//
	// if _, ok := err.(StackError); ok {
	// 	return err
	// }
	//
	// return &withStack{
	// 	err:   err,
	// 	stack: callers(3),
	// }
}

func (err *withStack) Stack() string {
	return "\n" + red + err.Error() + reset + "\n\n" + err.stack.String()
}

func (err *withStack) Unwrap() error {
	return err.err
}

func (err *withStack) Error() string {
	return err.Unwrap().Error()
}

func (err *withStack) Errorln() string {
	topFrame := frame((*err.stack)[0])
	return fmt.Sprintf(
		"%s:%d: \u001b[31m%s\u001B[0m\n", topFrame.file(),
		topFrame.line(), err.err,
	)
}

func (err *withStack) String() string {
	return err.Unwrap().Error()
}
