package history

import (
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

type fileAction struct {
	Timestamp  time.Time            `json:"timestamp"`
	ActionType types.FileActionType `json:"actionType"`

	OriginPath      string       `json:"originPath"`
	OriginId        types.FileId `json:"originId"`
	DestinationPath string       `json:"destinationPath"`
	DestinationId   types.FileId `json:"destinationId"`

	LifeId   types.LifetimeId `json:"lifeId"`
	LifeNext *fileAction      `json:"-" bson:"-"`
	LifePrev *fileAction      `json:"-" bson:"-"`
}

// func NewFileAction(actionType types.FileActionType) types.FileAction {
// 	return &fileAction{
// 		Timestamp:  time.Now(),
// 		ActionType: actionType,
// 	}
// }

func NewCreateAction(path string) types.FileAction {
	return &fileAction{
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: path,
		DestinationId:   types.SERV.FileTree.GenerateFileId(path),
	}
}

func NewMoveAction(originId, destinationId types.FileId) types.FileAction {
	return &fileAction{
		Timestamp:     time.Now(),
		ActionType:    FileCreate,
		OriginId:      originId,
		DestinationId: destinationId,
	}
}

func (fa *fileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

func (fa *fileAction) SetLifetimeId(lId types.LifetimeId) {
	fa.LifeId = lId
}

func (fa *fileAction) SetOriginPath(path string) {
	fa.OriginId = types.SERV.FileTree.GenerateFileId(path)
	fa.OriginPath = path
}

func (fa *fileAction) GetOriginPath() string {
	return fa.OriginPath
}

func (fa *fileAction) GetOriginId() types.FileId {
	return fa.OriginId
}

func (fa *fileAction) SetDestinationPath(path string) {
	fa.DestinationId = types.SERV.FileTree.GenerateFileId(path)
	fa.DestinationPath = path
}

func (fa *fileAction) GetDestinationPath() string {
	return fa.DestinationPath
}

func (fa *fileAction) GetDestinationId() types.FileId {
	return fa.DestinationId
}

func (fa *fileAction) SetActionType(actionType types.FileActionType) {
	fa.ActionType = actionType
}

func (fa *fileAction) GetActionType() types.FileActionType {
	return fa.ActionType
}
