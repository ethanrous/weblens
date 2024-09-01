package fileTree

import (
	"encoding/json"
	"time"

	"github.com/ethrousseau/weblens/internal/werror"
)

type FileActionType string

type FileAction struct {
	Timestamp  time.Time      `json:"timestamp" bson:"timestamp"`
	ActionType FileActionType `json:"actionType" bson:"actionType"`

	OriginPath string `json:"originPath" bson:"originPath,omitempty"`
	// OriginId        FileId `json:"originId" bson:"originId,omitempty"`
	DestinationPath string `json:"destinationPath" bson:"destinationPath,omitempty"`
	// DestinationId   FileId `json:"destinationId" bson:"destinationId,omitempty"`

	LifeId FileId `json:"lifeId" bson:"lifeId"`
	EventId FileEventId `json:"eventId" bson:"eventId"`

	Size     int64  `json:"size" bson:"size"`
	ParentId FileId `json:"parentId" bson:"ParentId"`
	ServerId string `json:"serverId" bson:"serverId"`

	LifeNext *FileAction `json:"-" bson:"-"`
	LifePrev *FileAction      `json:"-" bson:"-"`
	file     *WeblensFileImpl `bson:"-"`
}

func (fa *FileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

func (fa *FileAction) SetSize(size int64) {
	fa.Size = size
}

func (fa *FileAction) GetSize() int64 {
	return fa.Size
}

func (fa *FileAction) GetFile() *WeblensFileImpl {
	return fa.file
}

func (fa *FileAction) SetLifetimeId(lId FileId) {
	fa.LifeId = lId
}

func (fa *FileAction) GetLifetimeId() FileId {
	return fa.LifeId
}

func (fa *FileAction) GetOriginPath() string {
	return fa.OriginPath
}

// func (fa *FileAction) GetOriginId() FileId {
// 	return fa.OriginId
// }

func (fa *FileAction) GetDestinationPath() string {
	return fa.DestinationPath
}

// func (fa *FileAction) GetDestinationId() FileId {
// 	return fa.DestinationId
// }

func (fa *FileAction) SetActionType(actionType FileActionType) {
	fa.ActionType = actionType
}

func (fa *FileAction) GetActionType() FileActionType {
	return fa.ActionType
}

func (fa *FileAction) GetEventId() FileEventId {
	return fa.EventId
}

func (fa *FileAction) GetParentId() FileId {
	return fa.ParentId
}

func (fa *FileAction) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"timestamp":  fa.Timestamp.UnixMilli(),
		"actionType": fa.ActionType,
		"originPath": fa.OriginPath,
		// "originId":        fa.OriginId,
		"destinationPath": fa.DestinationPath,
		// "destinationId":   fa.DestinationId,
		"lifeId":   fa.LifeId,
		"eventId":  fa.EventId,
		"size":     fa.Size,
		"parentId": fa.ParentId,
	}

	return json.Marshal(data)
}

func (fa *FileAction) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return werror.WithStack(err)
	}

	fa.Timestamp = time.UnixMilli(int64(data["timestamp"].(float64)))
	fa.ActionType = FileActionType(data["actionType"].(string))
	fa.OriginPath = data["originPath"].(string)
	// fa.OriginId = FileId(data["originId"].(string))
	fa.DestinationPath = data["destinationPath"].(string)
	// fa.DestinationId = FileId(data["destinationId"].(string))
	fa.LifeId = FileId(data["lifeId"].(string))
	fa.EventId = FileEventId(data["eventId"].(string))
	fa.Size = int64(data["size"].(float64))
	fa.ParentId = FileId(data["parentId"].(string))

	return nil
}

const (
	FileCreate  FileActionType = "fileCreate"
	FileMove   FileActionType = "fileMove"
	FileWrite  FileActionType = "fileWrite"
	Backup     FileActionType = "backup"
	FileDelete FileActionType = "fileDelete"
	FileRestore FileActionType = "fileRestore"
)
