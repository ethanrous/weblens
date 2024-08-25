package werror

import (
	"fmt"
	"runtime"
	"strconv"
)

type base struct {
	msg string
}

type werr struct {
	error
	code int
	text string
	*stack
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

type stack []uintptr

func (s *stack) String() (stack string) {
	for _, pc := range *s {
		stack += fmt.Sprintf("%+v\n", Frame(pc))
	}
	return
}

type withStack struct {
	err   error
	stack *stack
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(StackError); ok {
		return err
	}

	if err == nil {
		return nil
	}
	return &withStack{
		err:   err,
		stack: callers(),
	}
}

func New(msg string) error {
	return base{
		msg: msg,
	}
}

func Errorf(format string, args ...any) StackError {
	return &withStack{
		err:   fmt.Errorf(format, args...),
		stack: callers(),
	}
}

func (err *withStack) Stack() string {
	return fmt.Sprintf("\u001b[31m%s\u001B[0m\n%s", err.err, err.stack.String())
}

func (err *withStack) Unwrap() error {
	return err.err
}

func (err *withStack) Error() string {
	topFrame := Frame((*err.stack)[0])
	return fmt.Sprintf(
		"%s:%d: \u001b[31m%s\u001B[0m\n", topFrame.file(),
		topFrame.line(), err.err,
	)
}

func (err base) Error() string {
	return err.msg
}

var NotImplemented = func(note string) error {
	return &withStack{
		err:   fmt.Errorf("not implemented: %s", note),
		stack: callers(),
	}
}

var _ StackError = (*withStack)(nil)

type StackError interface {
	Error() string
	Stack() string
}

type Frame uintptr

func (f Frame) pc() uintptr { return uintptr(f) - 1 }

func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func (f Frame) String() (str string) {
	return f.name() + "\n\t" + f.file() + ":" + strconv.Itoa(f.line())
}
