package types

import (
	"fmt"
	"strings"
	"time"

	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var ErrAlreadyInitialized = error2.WErrMsg("attempting to run an initialization routine for a second time")
var ErrServerNotInit = error2.WErrMsg("server is not initialized")

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
