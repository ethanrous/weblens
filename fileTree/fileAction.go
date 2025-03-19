package fileTree

import (
	"time"
)

type FileActionType = string

const (
	FileCreate     FileActionType = "fileCreate"
	FileMove       FileActionType = "fileMove"
	FileSizeChange FileActionType = "fileSizeChange"
	Backup         FileActionType = "backup"
	FileDelete     FileActionType = "fileDelete"
	FileRestore    FileActionType = "fileRestore"
)

type FileAction struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	file            *WeblensFileImpl `bson:"-"`
	ActionType      FileActionType   `json:"actionType" bson:"actionType"`
	OriginPath      string           `json:"originPath" bson:"originPath,omitempty"`
	DestinationPath string           `json:"destinationPath" bson:"destinationPath,omitempty"`
	LifeId          FileId           `json:"lifeId" bson:"lifeId"`
	EventId         FileEventId      `json:"eventId" bson:"eventId"`
	ParentId        FileId           `json:"parentId" bson:"parentId"`
	ServerId        string           `json:"serverId" bson:"serverId"`

	Size int64 `json:"size" bson:"size"`
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

func (fa *FileAction) GetDestinationPath() string {
	return fa.DestinationPath
}

func (fa *FileAction) GetRelevantPath() string {
	if fa.GetActionType() == FileDelete {
		return fa.OriginPath
	}
	return fa.DestinationPath
}

func (fa *FileAction) GetActionType() FileActionType {
	return fa.ActionType
}

func (fa *FileAction) GetParentId() FileId {
	return fa.ParentId
}
