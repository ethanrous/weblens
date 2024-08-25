package fileTree

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
)

type FileEventId string

type FileEvent struct {
	EventId    FileEventId   `bson:"_id"`
	Actions    []*FileAction `bson:"actions"`
	EventBegin time.Time     `bson:"eventBegin"`
	ServerId   string        `bson:"serverId"`

	journal     JournalService `bson:"-"`
	ActionsLock sync.Mutex     `bson:"-"`
}

// NewFileEvent returns a FileEvent, a container for multiple FileActions that occur due to the
// same event (move, delete, etc.)
// func NewFileEvent(journal JournalService) *FileEvent {
// 	return &FileEvent{
// 		EventId: FileEventId(primitive.NewObjectID().Hex()),
// 		EventBegin:  time.Now(),
// 		Actions:     []*FileAction{},
// 		journal: journal,
// 	}
// }

func (fe *FileEvent) GetEventId() FileEventId {
	return fe.EventId
}

func (fe *FileEvent) addAction(a *FileAction) {
	fe.ActionsLock.Lock()
	defer fe.ActionsLock.Unlock()

	fe.Actions = append(fe.Actions, a)
}

func (fe *FileEvent) GetActions() []*FileAction {
	return internal.SliceConvert[*FileAction](fe.Actions)
}

func (fe *FileEvent) NewCreateAction(file *WeblensFile) *FileAction {
	if fe.journal == nil {
		return nil
	}

	newAction := &FileAction{
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: file.GetPortablePath().ToPortable(),
		DestinationId: file.ID(),
		EventId:         fe.EventId,
		ParentId: file.GetParentId(),
		ServerId: fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewMoveAction(originId FileId, file *WeblensFile) *FileAction {
	if fe.journal == nil {
		return nil
	}

	lt := fe.journal.GetLifetimeByFileId(originId)
	if lt == nil {
		log.Error.Println("Cannot not find existing lifetime for originId", originId)
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		log.Error.Println("File previous destination does not match move origin")
	}

	newAction := &FileAction{
		Timestamp:       time.Now(),
		ActionType:      FileMove,
		OriginId:        latest.GetDestinationId(),
		OriginPath:      latest.GetDestinationPath(),
		DestinationId:   file.ID(),
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),
		ServerId: fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewDeleteAction(originId FileId) *FileAction {
	if fe.journal == nil {
		return nil
	}

	lt := fe.journal.GetLifetimeByFileId(originId)
	if lt == nil {
		log.ShowErr(werror.Errorf("Cannot not find existing lifetime for originId [%s]", originId))
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		log.Error.Println("File previous destination does not match move origin")
	}

	newAction := &FileAction{
		Timestamp:  time.Now(),
		ActionType: FileDelete,
		OriginId:   latest.GetDestinationId(),
		OriginPath: latest.GetDestinationPath(),
		EventId:    fe.EventId,
		ServerId: fe.ServerId,
	}

	fe.addAction(newAction)

	return newAction
}
