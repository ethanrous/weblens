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

	journal     Journal      `bson:"-"`
	hasher      Hasher       `bson:"-"`
	actionsLock sync.RWMutex `bson:"-"`

	// LoggedChan is used to signal that the event has been logged to the journal.
	// This is used to prevent actions on the same lifetime to be logged out of order.
	// LoggedChan does not get written to, it is only closed.
	LoggedChan chan struct{} `bson:"-"`
}

func (fe *FileEvent) addAction(a *FileAction) {
	fe.actionsLock.Lock()
	defer fe.actionsLock.Unlock()

	fe.Actions = append(fe.Actions, a)
}

func (fe *FileEvent) GetActions() []*FileAction {
	fe.actionsLock.RLock()
	defer fe.actionsLock.RUnlock()

	return fe.Actions
}

func (fe *FileEvent) NewCreateAction(file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	log.Trace.Printf("Building create action for [%s]", file.Filename())

	if !file.IsDir() && file.GetContentId() == "" {
		fe.hasher.Hash(file)
	}

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParentId(),
		ServerId:        fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) GetEventId() FileEventId {
	return fe.EventId
}

func (fe *FileEvent) Wait() {
	log.Trace.Printf("Waiting for event [%s] to be logged", fe.EventId)
	<-fe.LoggedChan
	log.Trace.Printf("Event [%s] logged", fe.EventId)
}

func (fe *FileEvent) NewMoveAction(lifeId FileId, file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	lt := fe.journal.Get(lifeId)
	if lt == nil {
		err := werror.Errorf("Cannot not find existing lifetime for %s", lifeId)
		log.ErrTrace(err)
		return nil
	}
	latest := lt.GetLatestAction()

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileMove,
		OriginPath:      latest.GetDestinationPath(),
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),
		ServerId:        fe.ServerId,

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
		err := werror.Errorf("Cannot not find existing lifetime for %s", lifeId)
		log.ErrTrace(err)
		return nil
	}

	for _, action := range fe.GetActions() {
		if action.LifeId == lifeId {
			panic("Got duplicate lifeId in file event")
		}
	}

	latest := lt.GetLatestAction()

	// if latest.GetDestinationId() != lifeId {
	// 	log.Error.Println("File previous destination does not match move origin")
	// }

	newAction := &FileAction{
		LifeId:     lifeId,
		Timestamp:  time.Now(),
		ActionType: FileDelete,
		OriginPath: latest.GetDestinationPath(),
		EventId:    fe.EventId,
		ServerId:   fe.ServerId,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewRestoreAction(file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	log.Trace.Printf("Building restore action for [%s]", file.Filename())

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileRestore,
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParentId(),
		ServerId:        fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewSizeChangeAction(file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		log.Trace.Println("Journal not set on size change action")
		return nil
	}

	log.Trace.Printf("Building size change action for [%s]", file.Filename())
	lt := fe.journal.Get(file.ID())
	if lt == nil {
		err := werror.Errorf("Cannot not find existing lifetime for %s", file.ID())
		log.ErrTrace(err)
		return nil
	}

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       time.Now(),
		ActionType:      FileSizeChange,
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParentId(),
		ServerId:        fe.ServerId,
		Size:            file.Size(),

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}
