package fileTree

import (
	"sync"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
)

type Lifetime struct {
	Id FileId `bson:"_id" json:"id"`
	// LiveFileId   FileId        `bson:"liveFileId" json:"liveFileId"`
	LiveFilePath string        `bson:"liveFilePath" json:"liveFilePath"`
	ContentId    string        `bson:"contentId,omitempty" json:"contentId,omitempty"`
	Actions      []*FileAction `bson:"actions" json:"actions"`
	ServerId     string        `bson:"serverId" json:"serverId"`

	actionsLock sync.RWMutex
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
		log.Error.Printf("No content file: %s", createAction.OriginPath)
		return nil, werror.Errorf("cannot create regular file lifetime without content id")
	}

	return &Lifetime{
		Id:           createAction.LifeId,
		LiveFilePath: createAction.GetDestinationPath(),
		Actions:      []*FileAction{createAction},
		ContentId:    createAction.file.GetContentId(),
		ServerId:     createAction.ServerId,
	}, nil
}

func (l *Lifetime) ID() FileId {
	return l.Id
}

func (l *Lifetime) Add(action *FileAction) {
	l.actionsLock.Lock()
	defer l.actionsLock.Unlock()

	action.SetLifetimeId(l.Id)
	l.Actions = append(l.Actions, action)
	// l.LiveFileId = action.GetDestinationId()
	l.LiveFilePath = action.GetDestinationPath()
}

// func (l *Lifetime) GetLatestFileId() FileId {
// 	return l.LiveFileId
// }

func (l *Lifetime) GetLatestFilePath() string {
	return l.LiveFilePath
}

func (l *Lifetime) GetLatestAction() *FileAction {
	return l.Actions[len(l.Actions)-1]
}

func (l *Lifetime) GetLatestSize() int64 {
	for _, a := range l.Actions {
		if a.Size != 0 {
			return a.Size
		}
	}

	return 0
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
	return internal.SliceConvert[*FileAction](l.Actions)
}
