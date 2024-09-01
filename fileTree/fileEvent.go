package fileTree

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
)

type FileEventId string

type FileEvent struct {
	EventId    FileEventId   `bson:"_id"`
	Actions    []*FileAction `bson:"actions"`
	EventBegin time.Time     `bson:"eventBegin"`
	ServerId   string        `bson:"serverId"`

	journal     JournalService `bson:"-"`
	ActionsLock sync.RWMutex `bson:"-"`
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
	fe.ActionsLock.RLock()
	defer fe.ActionsLock.RUnlock()

	return fe.Actions
}

func (fe *FileEvent) NewCreateAction(file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	newAction := &FileAction{
		LifeId:   file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId: file.GetParentId(),
		ServerId: fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewMoveAction(lifeId FileId, file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	lt := fe.journal.Get(lifeId)
	if lt == nil {
		log.Error.Println("Cannot not find existing lifetime for lifeId", lifeId)
		return nil
	}
	latest := lt.GetLatestAction()

	newAction := &FileAction{
		LifeId:   file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileMove,
		OriginPath:      latest.GetDestinationPath(),
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),
		ServerId: fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewDeleteAction(lifeId FileId) *FileAction {
	if fe.journal == nil {
		return nil
	}

	lt := fe.journal.Get(lifeId)
	if lt == nil {
		log.ShowErr(werror.Errorf("Cannot not find existing lifetime for lifeId [%s]", lifeId))
		return nil
	}
	latest := lt.GetLatestAction()

	// if latest.GetDestinationId() != lifeId {
	// 	log.Error.Println("File previous destination does not match move origin")
	// }

	newAction := &FileAction{
		LifeId:   lifeId,
		Timestamp:  time.Now(),
		ActionType: FileDelete,
		OriginPath: latest.GetDestinationPath(),
		EventId:    fe.EventId,
		ServerId: fe.ServerId,
	}

	fe.addAction(newAction)

	return newAction
}
