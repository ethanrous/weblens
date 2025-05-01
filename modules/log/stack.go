package log

import "github.com/pkg/errors"

var (
	StackSourceFileName     = "source"
	StackSourceLineName     = "line"
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

func frameField(f errors.Frame, s *state, c rune) string {
	f.Format(s, c)
	return string(s.b)
}

func MarshalStack(err error) any {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

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
