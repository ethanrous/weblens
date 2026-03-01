package wlog

import (
	"fmt"

	"github.com/ethanrous/weblens/modules/wlerrors"
)

var (
	// StackSourceFileName is the key for the source file name in stack traces.
	StackSourceFileName = "source"
	// StackSourceLineName is the key for the line number in stack traces.
	StackSourceLineName = "line"
	// StackSourceFunctionName is the key for the function name in stack traces.
	StackSourceFunctionName = "func"
)

type state struct {
	b []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b

	return len(b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	switch c {
	case '+':
		return true
	default:
		return false
	}
}

func frameField(f wlerrors.Frame, s *state, c rune) string {
	f.Format(s, c)

	return string(s.b)
}

type stackTracer interface {
	StackTrace() wlerrors.StackTrace
}

type stopper interface {
	Stop() bool
}

// MarshalStack extracts and marshals the stack trace from an error.
func MarshalStack(err error) any {
	s := &state{}

	var sterr stackTracer

	var ok bool

	for err != nil {
		var tmpStacker stackTracer

		tmpStacker, ok = err.(stackTracer)
		if ok {
			tmpStacker.StackTrace()[0].Format(s, 'n')

			if string(s.b) == "init" {
				break
			}

			sterr = tmpStacker

			if stop, ok := err.(stopper); ok && stop.Stop() {
				break
			}
		}

		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}

		err = u.Unwrap()
	}

	if sterr == nil {
		return nil
	}

	st := sterr.StackTrace()
	out := make([]map[string]string, 0, len(st))

	for _, frame := range st {
		out = append(out, map[string]string{
			StackSourceFileName:     frameField(frame, s, 's'),
			StackSourceLineName:     frameField(frame, s, 'd'),
			StackSourceFunctionName: frameField(frame, s, 'n'),
		})
	}

	return out
}

// PrintStackTrace captures and prints the current stack trace.
func PrintStackTrace() {
	err := wlerrors.New("stack trace")
	st := MarshalStack(err)

	stm, ok := st.([]map[string]string)
	if !ok {
		fmt.Println("Could not marshal stack trace", st)

		return
	}

	if len(stm) == 0 {
		fmt.Println("No stack trace available")
	}

	stma := make([]any, 0, len(stm)-1)

	for i, frame := range stm {
		if i == 0 {
			continue
		}

		stma = append(stma, frame)
	}

	stkStr := formatStack(stma)
	GlobalLogger().Log().Msg(stkStr)
}
