package types

import (
	"time"
)

type LifetimeId string
type FileActionType string
type FileEventId string

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
	AddAction(action FileAction)
	GetActions() []FileAction
	// UnmarshalBSON([]byte) error
	// UnmarshalBSONValue(t bsontype.Type, value []byte) error
}

type FileAction interface {
	GetContentId() ContentId

	GetOriginPath() string
	SetOriginPath(path string)
	GetOriginId() FileId

	GetDestinationPath() string
	SetDestinationPath(path string)
	GetDestinationId() FileId

	SetActionType(action FileActionType)
	GetActionType() FileActionType

	GetTimestamp() time.Time
}

type Lifetime interface {
	GetFileId() FileId
	GetContentId() ContentId
}
