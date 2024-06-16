package types

import (
	"time"
)

type JournalService interface {
	WatchFolder(f WeblensFile) error

	LogEvent(fe FileEvent) error

	JournalWorker()
	FileWatcher()
}

type FileActionType string

type FileEventId string

// FileEvent is a group of FileActions that take place at the same time
type FileEvent interface {
	GetEventId() FileEventId
	AddAction(action FileAction)
	GetActions() []FileAction
	// UnmarshalBSON([]byte) error
	// UnmarshalBSONValue(t bsontype.Type, value []byte) error
}

type LifetimeId string

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
