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
	Timestamp  time.Time            `json:"timestamp"`
	ActionType types.FileActionType `json:"actionType"`

	OriginPath      string       `json:"originPath"`
	OriginId        types.FileId `json:"originId"`
	DestinationPath string       `json:"destinationPath"`
	DestinationId   types.FileId `json:"destinationId"`

	LifeId  types.LifetimeId  `json:"lifeId"`
	EventId types.FileEventId `json:"eventId"`

	LifeNext *FileAction `json:"-" bson:"-"`
	LifePrev *FileAction `json:"-" bson:"-"`
}

// func NewFileAction(actionType types.FileActionType) types.FileAction {
// 	return &FileAction{
// 		Timestamp:  time.Now(),
// 		ActionType: actionType,
// 	}
// }

func (fa *FileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

func (fa *FileAction) SetLifetimeId(lId types.LifetimeId) {
	fa.LifeId = lId
}

func (fa *FileAction) SetOriginPath(path string) {
	fa.OriginPath = path
	fa.OriginId = types.SERV.FileTree.GenerateFileId(path)
}

func (fa *FileAction) GetOriginPath() string {
	return fa.OriginPath
}

func (fa *FileAction) GetOriginId() types.FileId {
	return fa.OriginId
}

func (fa *FileAction) SetDestinationPath(path string) {
	fa.DestinationPath = path
	fa.DestinationId = types.SERV.FileTree.GenerateFileId(path)
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

func (fa *FileAction) MarshalJSON() ([]byte, error) {
	data := map[string]any{}
	data["timestamp"] = fa.Timestamp.Unix()
	data["actionType"] = fa.ActionType
	data["originPath"] = fa.OriginPath
	data["originId"] = fa.OriginId
	data["destinationPath"] = fa.DestinationPath
	data["destinationId"] = fa.DestinationId
	data["lifeId"] = fa.LifeId
	data["eventId"] = fa.EventId

	return json.Marshal(data)
}
