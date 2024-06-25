package types

import (
	"time"
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

type JournalService interface {
	BaseService[LifetimeId, []FileAction]

	WatchFolder(f WeblensFile) error

	LogEvent(fe FileEvent) error

	JournalWorker()
	FileWatcher()
	GetActiveLifetimes() []Lifetime
}

// FileEvent is a group of FileActions that take place at the same time
type FileEvent interface {
	GetEventId() FileEventId
	// addAction(action FileAction)
	GetActions() []FileAction

	NewCreateAction(file WeblensFile) FileAction
	NewMoveAction(originId, destinationId FileId) FileAction
	// UnmarshalBSON([]byte) error
	// UnmarshalBSONValue(t bsontype.Type, value []byte) error
}

type FileAction interface {
	GetOriginPath() string
	SetOriginPath(path string)
	GetOriginId() FileId

	GetDestinationPath() string
	SetDestinationPath(path string)
	GetDestinationId() FileId

	SetActionType(action FileActionType)
	GetActionType() FileActionType

	SetLifetimeId(LifetimeId)

	GetTimestamp() time.Time
	GetEventId() FileEventId
}

type Lifetime interface {
	ID() LifetimeId
	Add(FileAction)
	GetLatestFileId() FileId
	GetContentId() ContentId
	IsLive() bool
	GetActions() []FileAction
}
