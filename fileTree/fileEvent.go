package fileTree

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileEventId = string

type FileEvent struct {
	EventBegin time.Time `bson:"eventBegin"`

	journal Journal `bson:"-"`
	hasher  Hasher  `bson:"-"`

	EventId  FileEventId `bson:"_id"`
	ServerId string      `bson:"serverId"`

	Actions     []*FileAction `bson:"actions"`
	actionsLock sync.RWMutex  `bson:"-"`

	// LoggedChan is used to signal that the event has been logged to the journal.
	// This is used to prevent actions on the same lifetime to be logged out of order.
	// LoggedChan does not get written to, it is only closed.
	LoggedChan chan struct{} `bson:"-"`
	Logged     atomic.Bool   `bson:"-"`
}

func NewFileEvent(journal Journal, serverId string, hasher Hasher) *FileEvent {
	return &FileEvent{
		EventBegin: time.Now(),
		EventId:    FileEventId(primitive.NewObjectID().Hex()),
		ServerId:   serverId,
		hasher:     hasher,
		journal:    journal,

		Actions:     make([]*FileAction, 0),
		actionsLock: sync.RWMutex{},

		LoggedChan: make(chan struct{}),
	}
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

	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building create action for [%s]", file.GetPortablePath()) })

	if !file.IsDir() && file.GetContentId() == "" {
		err := fe.hasher.Hash(file)
		if err != nil {
			log.Error().Stack().Err(werror.WithStack(err)).Msg("")
			return nil
		}
	}

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       fe.EventBegin,
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

func (fe *FileEvent) Wait() error {
	if fe == nil || fe.LoggedChan == nil {
		return werror.Errorf("Cannot wait on nil event")
	}
	if fe.Logged.Load() {
		return werror.Errorf("Event already logged")
	}

	<-fe.LoggedChan
	return nil
}

func (fe *FileEvent) SetLogged() {
	if fe == nil {
		log.Error().Stack().Err(werror.Errorf("Cannot set logged on nil event")).Msg("")
		return
	}
	if fe.Logged.Load() {
		log.Error().Stack().Err(werror.Errorf("Event [%s] already logged", fe.EventId)).Msg("")
		return
	}

	fe.Logged.Store(true)
	close(fe.LoggedChan)
}

func (fe *FileEvent) NewMoveAction(lifeId FileId, file *WeblensFileImpl) (*FileAction, error) {
	if fe.journal == nil {
		return nil, errors.Wrap(werror.ErrNoJournal, "journal not set on move action")
	}

	fe.journal.Flush()

	lt := fe.journal.Get(lifeId)
	if lt == nil {
		return nil, errors.Wrapf(werror.ErrNoLifetime, "Moving [%s] to [%s]", lifeId, file.GetPortablePath())
	}
	latest := lt.GetLatestAction()

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       fe.EventBegin,
		ActionType:      FileMove,
		OriginPath:      latest.GetDestinationPath(),
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),
		ServerId:        fe.ServerId,

		file: file,
	}

	fe.addAction(newAction)

	return newAction, nil
}

func (fe *FileEvent) NewDeleteAction(lifeId FileId) (*FileAction, error) {
	if fe.journal == nil {
		return nil, werror.Errorf("Journal not set on delete action")
	}

	fe.journal.Flush()

	lt := fe.journal.Get(lifeId)
	if lt == nil {
		return nil, errors.Wrapf(werror.ErrNoLifetime, "Deleting [%s]", lifeId)
	}

	for _, action := range fe.GetActions() {
		if action.LifeId == lifeId {
			err := werror.Errorf("Got duplicate lifeId in file event")
			return nil, err
		}
	}

	latest := lt.GetLatestAction()

	newAction := &FileAction{
		LifeId:     lifeId,
		Timestamp:  fe.EventBegin,
		ActionType: FileDelete,
		OriginPath: latest.GetDestinationPath(),
		EventId:    fe.EventId,
		ServerId:   fe.ServerId,
	}

	fe.addAction(newAction)

	return newAction, nil
}

func (fe *FileEvent) NewRestoreAction(file *WeblensFileImpl) *FileAction {
	if fe.journal == nil {
		return nil
	}

	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building restore action for [%s]", file.Filename()) })

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       fe.EventBegin,
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
		log.Trace().Msg("Journal not set on size change action")
		return nil
	}

	fe.journal.Flush()

	for _, action := range fe.GetActions() {
		if action.LifeId == file.ID() {
			action.Size = file.Size()
			return nil
		}
	}

	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building size change action for [%s]", file.Filename()) })
	lt := fe.journal.Get(file.ID())
	if lt == nil {
		err := werror.Errorf("Cannot find existing lifetime for %s", file.ID())
		log.Error().Stack().Err(err).Msg("")
		return nil
	}

	newAction := &FileAction{
		LifeId:          file.ID(),
		Timestamp:       fe.EventBegin,
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
