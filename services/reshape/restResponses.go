package reshape

// FIXME: This file is a mess, clean it up, move all of these structs to the structs package and all of the functions
// to their own files.

import (
	"encoding/json"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
)

type Option[T any] struct {
	value T
	set   bool
} // @name

func NewOption[T any](value T) Option[T] {
	return Option[T]{value: value, set: true}
}

func (o Option[T]) Set(v T) Option[T] {
	o.value = v
	o.set = true

	return o
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.set {
		return json.Marshal(o.value)
	}
	return nil, nil
}

type WeblensErrorInfo struct {
	Error string `json:"error"`
} // @name ErrorInfo

type DispatchInfo struct {
	TaskId string `json:"taskId"`
} // @name DispatchInfo

func ApiKeyToApiKeyInfo(k auth.Token) structs.ApiKeyInfo {
	return structs.ApiKeyInfo{
		Id:           k.Id.Hex(),
		Name:         k.Nickname,
		Key:          string(k.Token[:]),
		Owner:        k.Owner,
		CreatedTime:  k.CreatedTime.UnixMilli(),
		RemoteUsing:  k.RemoteUsing,
		CreatedBy:    k.CreatedBy,
		LastUsedTime: k.LastUsed.UnixMilli(),
	}
}
