package reshape

// FIXME: This file is a mess, clean it up, move all of these structs to the structs package and all of the functions
// to their own files.

import (
	"encoding/json"

	"github.com/ethanrous/weblens/models"
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

type TakeoutInfo struct {
	TakeoutId string `json:"takeoutId"`
	TaskId    string `json:"taskId"`
	Filename  string `json:"filename"`
	Single    bool   `json:"single"`
} // @name TakeoutInfo

type DispatchInfo struct {
	TaskId string `json:"taskId"`
} // @name DispatchInfo

type ApiKeyInfo struct {
	Id           string               `json:"id" validate:"required"`
	Name         string               `json:"name" validate:"required"`
	Key          models.WeblensApiKey `json:"key" validate:"required"`
	Owner        string               `json:"owner" validate:"required"`
	RemoteUsing  models.InstanceId    `json:"remoteUsing" validate:"required"`
	CreatedBy    models.InstanceId    `json:"createdBy" validate:"required"`
	CreatedTime  int64                `json:"createdTime" validate:"required"`
	LastUsedTime int64                `json:"lastUsedTime" validate:"required"`
} // @name ApiKeyInfo

func ApiKeyToApiKeyInfo(k models.ApiKey) ApiKeyInfo {
	return ApiKeyInfo{
		Id:           k.Id.Hex(),
		Name:         k.Name,
		Key:          k.Key,
		Owner:        k.Owner,
		CreatedTime:  k.CreatedTime.UnixMilli(),
		RemoteUsing:  k.RemoteUsing,
		CreatedBy:    k.CreatedBy,
		LastUsedTime: k.LastUsed.UnixMilli(),
	}
}

type NewUploadInfo struct {
	UploadId string `json:"uploadId"`
} // @name NewUploadInfo

type NewFileInfo struct {
	FileId string `json:"fileId"`
} // @name NewFileInfo

type NewFilesInfo struct {
	FileIds []string `json:"fileIds"`
} // @name NewFilesInfo

type RestoreFilesInfo struct {
	NewParentId string `json:"newParentId"`
} //	@name	RestoreFilesInfo
