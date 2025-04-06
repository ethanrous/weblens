package history

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/modules/fs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileEvent struct {
	EventBegin time.Time `bson:"eventBegin"`

	EventId string `bson:"_id"`
	TowerId string `bson:"serverId"`

	Actions     []*FileAction `bson:"actions"`
	actionsLock sync.RWMutex  `bson:"-"`

	// LoggedChan is used to signal that the event has been logged to the journal.
	// This is used to prevent actions on the same lifetime to be logged out of order.
	// LoggedChan does not get written to, it is only closed.
	LoggedChan chan struct{} `bson:"-"`
	Logged     atomic.Bool   `bson:"-"`
}

func NewFileEvent(serverId string) *FileEvent {
	return &FileEvent{
		EventBegin: time.Now(),
		EventId:    string(primitive.NewObjectID().Hex()),
		TowerId:    serverId,

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

func (fe *FileEvent) NewCreateAction(filepath fs.Filepath) *FileAction {
	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building create action for [%s]", filepath) })

	// if !file.IsDir() && file.GetContentId() == "" {
	// 	err := fe.hasher.Hash(file)
	// 	if err != nil {
	// 		log.Error().Stack().Err(errors.WithStack(err)).Msg("")
	// 		return nil
	// 	}
	// }

	newAction := &FileAction{
		Filepath:   filepath,
		Timestamp:  fe.EventBegin,
		ActionType: FileCreate,
		EventId:    fe.EventId,
		TowerId:    fe.TowerId,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) GetEventId() string {
	return fe.EventId
}

func (fe *FileEvent) Wait() error {
	if fe == nil || fe.LoggedChan == nil {
		return errors.Errorf("Cannot wait on nil event")
	}
	if fe.Logged.Load() {
		return errors.Errorf("Event already logged")
	}

	<-fe.LoggedChan
	return nil
}

func (fe *FileEvent) SetLogged() {
	if fe == nil {
		log.Error().Stack().Err(errors.Errorf("Cannot set logged on nil event")).Msg("")
		return
	}
	if fe.Logged.Load() {
		log.Error().Stack().Err(errors.Errorf("Event [%s] already logged", fe.EventId)).Msg("")
		return
	}

	fe.Logged.Store(true)
	close(fe.LoggedChan)
}

func (fe *FileEvent) NewMoveAction(originPath fs.Filepath, destinationPath fs.Filepath) (*FileAction, error) {
	// lt := fe.journal.Get(lifeId)
	// if lt == nil {
	// 	return nil, errors.Wrapf(errors.ErrNoLifetime, "Moving [%s] to [%s]", lifeId, file.GetPortablePath())
	// }
	// latest := lt.GetLatestAction()

	newAction := &FileAction{
		Timestamp:       fe.EventBegin,
		ActionType:      FileMove,
		OriginPath:      originPath,
		DestinationPath: destinationPath,
		EventId:         fe.EventId,
		TowerId:         fe.TowerId,
	}

	fe.addAction(newAction)

	return newAction, nil
}

func (fe *FileEvent) NewDeleteAction(filepath fs.Filepath) (*FileAction, error) {
	newAction := &FileAction{
		Filepath:   filepath,
		Timestamp:  fe.EventBegin,
		ActionType: FileDelete,
		EventId:    fe.EventId,
		TowerId:    fe.TowerId,
	}

	fe.addAction(newAction)

	return newAction, nil
}

func (fe *FileEvent) NewRestoreAction(filepath fs.Filepath, newSize int64) *FileAction {
	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building restore action for [%s]", filepath) })

	newAction := &FileAction{
		Filepath:   filepath,
		Timestamp:  fe.EventBegin,
		ActionType: FileRestore,
		EventId:    fe.EventId,
		TowerId:    fe.TowerId,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewSizeChangeAction(filepath fs.Filepath, newSize int64) *FileAction {
	for _, action := range fe.GetActions() {
		if action.Filepath == filepath {
			action.Size = newSize
			return nil
		}
	}

	log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Building size change action for [%s]", filepath) })

	newAction := &FileAction{
		Filepath:   filepath,
		Timestamp:  fe.EventBegin,
		ActionType: FileSizeChange,
		EventId:    fe.EventId,
		TowerId:    fe.TowerId,
		Size:       newSize,
	}

	fe.addAction(newAction)

	return newAction
}
