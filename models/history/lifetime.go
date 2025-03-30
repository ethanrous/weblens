package history

import (
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Lifetime struct {
	Id        string        `bson:"_id" json:"id"`
	ContentId string        `bson:"contentId,omitempty" json:"contentId,omitempty"`
	ServerId  string        `bson:"serverId" json:"serverId"`
	Actions   []*FileAction `bson:"actions" json:"actions"`

	actionsLock sync.RWMutex
	IsDir       bool `bson:"isDir" json:"isDir"`
}

func NewLifetime(createAction *FileAction) (*Lifetime, error) {
	actionType := createAction.GetActionType()
	if actionType != FileCreate && actionType != FileRestore {
		return nil, werror.Errorf("First Lifetime action must be of type FileCreate or FileRestore")
	}

	if createAction.file == nil {
		return nil, werror.Errorf("Could not find file to create lifetime with")
	}

	if !createAction.file.IsDir() && createAction.file.GetContentId() == "" && createAction.file.Size() != 0 {
		return nil, errors.Errorf("trying to create a regular file lifetime without expected content id: [%s]", createAction.OriginPath)
	}

	return &Lifetime{
		Id:        createAction.LifeId,
		Actions:   []*FileAction{createAction},
		IsDir:     createAction.file.IsDir(),
		ContentId: createAction.file.GetContentId(),
		ServerId:  createAction.ServerId,
	}, nil
}

func (l *Lifetime) ID() string {
	return l.Id
}

func (l *Lifetime) GetIsDir() bool {
	return l.IsDir
}

func (l *Lifetime) Add(action *FileAction) {
	l.actionsLock.Lock()
	defer l.actionsLock.Unlock()

	action.SetLifetimeId(l.Id)
	l.Actions = append(l.Actions, action)
}

func (l *Lifetime) GetLatestAction() *FileAction {
	return l.Actions[len(l.Actions)-1]
}

func (l *Lifetime) GetLatestSize() int64 {
	if len(l.Actions) == 0 {
		return 0
	}

	return l.Actions[len(l.Actions)-1].Size
}

func (l *Lifetime) GetLatestPath() fs.Filepath {
	i := len(l.Actions) - 1
	for i >= 0 {
		if l.Actions[i].ActionType == FileDelete {
			return fs.Filepath{}
		}
		if l.Actions[i].DestinationPath != "" {
			return fs.ParsePortable(l.Actions[i].DestinationPath)
		}
		i--
	}

	return fs.Filepath{}
}

// GetLatestMove returns the most recent move or create action in the lifetime. Ideally,
// this will show the current path of the file
func (l *Lifetime) GetLatestMove() *FileAction {
	if len(l.Actions) == 0 {
		return nil
	}

	i := len(l.Actions) - 1
	for i >= 0 {
		if l.Actions[i].ActionType != FileSizeChange {
			return l.Actions[i]
		}
		i--
	}

	return nil
}

func (l *Lifetime) HasEvent(eventId FileEventId) bool {
	return slices.ContainsFunc(
		l.Actions, func(a *FileAction) bool {
			return a.EventId == eventId
		},
	)
}

func (l *Lifetime) GetContentId() string {
	return l.ContentId
}

func (l *Lifetime) SetContentId(cId string) {
	l.ContentId = cId
}

// IsLive returns a boolean representing if this Lifetime shows a file
// currently on the real filesystem, and has not been deleted.
func (l *Lifetime) IsLive() bool {
	// If the most recent action has no destination, the file was removed
	return l.Actions[len(l.Actions)-1].DestinationPath != ""
}

func (l *Lifetime) GetActions() []*FileAction {
	l.actionsLock.RLock()
	defer l.actionsLock.RUnlock()
	return l.Actions
}

func LifetimeSorter(a, b *Lifetime) int {
	// Sort lifetimes by their most recent move time
	aLatestMove := a.GetLatestMove()
	if aLatestMove == nil {
		log.Error().Msgf("LifetimeSorter: a is nil for %s", a.Id)
		return 1
	}

	bLatestMove := b.GetLatestMove()
	if bLatestMove == nil {
		log.Error().Msgf("LifetimeSorter: b is nil for %s", b.Id)
		return -1
	}

	timeDiff := aLatestMove.GetTimestamp().Sub(bLatestMove.GetTimestamp())
	if timeDiff != 0 {
		return int(timeDiff)
	}

	// If the creation time is the same, sort by the path length. This is to ensure parent directories are created before their children.
	return len(aLatestMove.DestinationPath) - len(bLatestMove.DestinationPath)
}
