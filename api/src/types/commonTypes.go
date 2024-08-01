package types

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// type WeblensError interface {
// 	Error() string
// }

type WeblensError struct {
	err        error
	sourceFile string
	sourceLine int
	trace      string
}

func WeblensErrorFromError(err error) WeblensError {
	wlErr, ok := err.(WeblensError)
	if !ok {
		return NewWeblensError(err.Error())
	}
	return wlErr
}

func WeblensErrorMsg(err string) WeblensError {
	return NewWeblensError(err)
}

func NewWeblensError(err string) WeblensError {
	_, filename, line, _ := runtime.Caller(2)
	buf := make([]byte, 1<<16)

	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	return WeblensError{errors.New(err), filepath.Base(filename), line, string(buf)}
}

func (e WeblensError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.sourceFile, e.sourceLine, e.err)
}

func (e WeblensError) ErrorTrace() string {
	return fmt.Sprintf("%s:%d: %s\n%s", e.sourceFile, e.sourceLine, e.err, e.trace)
}

func (e WeblensError) GetSourceFile() string {
	return e.sourceFile
}

func (e WeblensError) GetSourceLine() int {
	return e.sourceLine
}

var ErrAlreadyInitialized = WeblensErrorMsg("attempting to run an initialization routine for a second time")
var ErrServerNotInit = WeblensErrorMsg("server is not initialized")

type SafeTime time.Time

func (t SafeTime) MarshalJSON() ([]byte, error) {
	// do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05.999999999 -0700 MST"))
	return []byte(stamp), nil
}

func (t *SafeTime) UnmarshalJSON(data []byte) error {
	// timeStr := strings.Trim(string(data), "\\")
	timeStr := strings.Trim(string(data), "\"")
	// timeStr := string(data)[1 : len(data)-1]
	realTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeStr)
	if err != nil {
		return err
	}
	*t = SafeTime(realTime)
	return nil
}

func (t SafeTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(time.Time(t))
}

func (t *SafeTime) UnmarshalBSONValue(bt bsontype.Type, data []byte) error {
	rv := bson.RawValue{
		Type:  bt,
		Value: data,
	}

	var res time.Time
	if err := rv.Unmarshal(&res); err != nil {
		return err
	}
	*t = SafeTime(res)

	return nil
}

func FromSafeTimeStr(timeStr string) (t time.Time, err error) {
	ti, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeStr)
	if err != nil {
		return
	}
	return ti, nil
}
