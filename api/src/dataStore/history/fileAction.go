package history

import (
	"encoding/json"
	"time"

	"github.com/ethrousseau/weblens/api/types"
)

const (
	FileCreate  types.FileActionType = "fileCreate"
	FileRestore types.FileActionType = "fileRestore"
	FileMove    types.FileActionType = "fileMove"
	FileDelete  types.FileActionType = "fileDelete"
	FileWrite   types.FileActionType = "fileWrite"
	Backup      types.FileActionType = "backup"
)

type FileAction struct {
	Timestamp  time.Time            `json:"timestamp" bson:"timestamp"`
	ActionType types.FileActionType `json:"actionType" bson:"actionType"`

	OriginPath      string       `json:"originPath" bson:"originPath,omitempty"`
	OriginId        types.FileId `json:"originId" bson:"originId,omitempty"`
	DestinationPath string       `json:"destinationPath" bson:"destinationPath,omitempty"`
	DestinationId   types.FileId `json:"destinationId" bson:"destinationId,omitempty"`

	LifeId  types.LifetimeId  `json:"lifeId" bson:"lifeId"`
	EventId types.FileEventId `json:"eventId" bson:"eventId"`

	Size     int64        `json:"size" bson:"size"`
	ParentId types.FileId `json:"parentId" bson:"ParentId"`

	LifeNext *FileAction       `json:"-" bson:"-"`
	LifePrev *FileAction       `json:"-" bson:"-"`
	file     types.WeblensFile `bson:"-"`
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

func (fa *FileAction) GetFile() types.WeblensFile {
	return fa.file
}

func (fa *FileAction) SetLifetimeId(lId types.LifetimeId) {
	fa.LifeId = lId
}

func (fa *FileAction) GetLifetimeId() types.LifetimeId {
	return fa.LifeId
}

func (fa *FileAction) GetOriginPath() string {
	return fa.OriginPath
}

func (fa *FileAction) GetOriginId() types.FileId {
	return fa.OriginId
}

func (fa *FileAction) GetDestinationPath() string {
	return fa.DestinationPath
}

func (fa *FileAction) GetDestinationId() types.FileId {
	return fa.DestinationId
}

func (fa *FileAction) SetActionType(actionType types.FileActionType) {
	fa.ActionType = actionType
}

func (fa *FileAction) GetActionType() types.FileActionType {
	return fa.ActionType
}

func (fa *FileAction) GetEventId() types.FileEventId {
	return fa.EventId
}

func (fa *FileAction) GetParentId() types.FileId {
	return fa.ParentId
}

func (fa *FileAction) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"timestamp":       fa.Timestamp.UnixMilli(),
		"actionType":      fa.ActionType,
		"originPath":      fa.OriginPath,
		"originId":        fa.OriginId,
		"destinationPath": fa.DestinationPath,
		"destinationId":   fa.DestinationId,
		"lifeId":          fa.LifeId,
		"eventId":         fa.EventId,
		"size":            fa.Size,
		"parentId":        fa.ParentId,
	}

	return json.Marshal(data)
}

func (fa *FileAction) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}

	fa.Timestamp = time.UnixMilli(int64(data["timestamp"].(float64)))
	fa.ActionType = types.FileActionType(data["actionType"].(string))
	fa.OriginPath = data["originPath"].(string)
	fa.OriginId = types.FileId(data["originId"].(string))
	fa.DestinationPath = data["destinationPath"].(string)
	fa.DestinationId = types.FileId(data["destinationId"].(string))
	fa.LifeId = types.LifetimeId(data["lifeId"].(string))
	fa.EventId = types.FileEventId(data["eventId"].(string))
	fa.Size = int64(data["size"].(float64))
	fa.ParentId = types.FileId(data["parentId"].(string))

	return nil
}
