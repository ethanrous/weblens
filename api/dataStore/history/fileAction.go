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
	ContentId  types.ContentId      `json:"contentId"`

	OriginPath      string       `json:"originPath"`
	OriginId        types.FileId `json:"originId"`
	DestinationPath string       `json:"destinationPath"`
	DestinationId   types.FileId `json:"destinationId"`

	LifeId   types.LifetimeId `json:"lifeId"`
	LifeNext *fileAction      `json:"-"`
	LifePrev *fileAction      `json:"-"`
}

func NewFileAction(contentId types.ContentId, actionType types.FileActionType) types.FileAction {
	return &fileAction{
		Timestamp:  time.Now(),
		ActionType: actionType,
		ContentId:  contentId,
	}
}

func NewCreateEntry(path string, contentId types.ContentId) types.FileAction {
	return &fileAction{
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		ContentId:       contentId,
		DestinationPath: path,
		DestinationId:   hc.fileTree.GenerateFileId(path),
	}
}

func (fa *fileAction) GetContentId() types.ContentId {
	return fa.ContentId
}

func (fa *fileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

func (fa *fileAction) SetOriginPath(path string) {
	fa.OriginId = hc.fileTree.GenerateFileId(path)
	fa.OriginPath = path
}

func (fa *fileAction) GetOriginPath() string {
	return fa.OriginPath
}

func (fa *fileAction) GetOriginId() types.FileId {
	return fa.OriginId
}

func (fa *fileAction) SetDestinationPath(path string) {
	fa.DestinationId = hc.fileTree.GenerateFileId(path)
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
