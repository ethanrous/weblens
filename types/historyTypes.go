package types

import (
	"time"

	"github.com/ethrousseau/weblens/api"
)

type LifetimeId string
type FileActionType string
type FileEventId string

const (
	FileCreate  FileActionType = "fileCreate"
	FileRestore FileActionType = "fileRestore"
	FileMove    FileActionType = "fileMove"
	FileDelete  FileActionType = "fileDelete"
	FileWrite   FileActionType = "fileWrite"
	FileBackup  FileActionType = "backup"
)

// FileEvent is a group of FileActions that take place at the same time
type FileEvent interface {
	GetEventId() FileEventId
	GetActions() []FileAction

	NewCreateAction(file WeblensFile) FileAction
	NewMoveAction(originId FileId, file WeblensFile) FileAction
	NewDeleteAction(originId FileId) FileAction
}

type FileAction interface {
	SetSize(size int64)
	GetSize() int64

	GetOriginPath() string
	GetOriginId() FileId

	GetDestinationPath() string
	GetDestinationId() FileId

	SetActionType(action FileActionType)
	GetActionType() FileActionType

	GetLifetimeId() LifetimeId
	SetLifetimeId(LifetimeId)

	GetTimestamp() time.Time
	GetEventId() FileEventId

	GetParentId() FileId

	GetFile() WeblensFile
}

type Lifetime interface {
	ID() LifetimeId
	Add(FileAction)
	GetLatestFileId() FileId
	GetLatestAction() FileAction
	GetContentId() weblens.ContentId
	SetContentId(weblens.ContentId)
	IsLive() bool
	GetActions() []FileAction
}
