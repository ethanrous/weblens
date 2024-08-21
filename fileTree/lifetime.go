package fileTree

import (
	"sync"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LifetimeId string

type Lifetime struct {
	Id         LifetimeId        `bson:"_id" json:"id"`
	LiveFileId FileId            `bson:"liveFileId" json:"liveFileId"`
	ContentId  weblens.ContentId `bson:"contentId,omitempty" json:"contentId,omitempty"`
	Actions    []*FileAction     `bson:"actions" json:"actions"`
	ServerId   types.InstanceId  `bson:"serverId" json:"serverId"`

	actionsLock sync.RWMutex
}

func NewLifetime(id LifetimeId, createAction types.FileAction) (*Lifetime, error) {
	if createAction.GetActionType() != FileCreate {
		return nil, werror.NewWeblensError("First Lifetime action must be of type FileCreate")
	}

	if id == "" {
		id = LifetimeId(primitive.NewObjectID().Hex())
	}

	createAction.SetLifetimeId(id)

	file := types.SERV.FileTree.Get(createAction.GetDestinationId())
	if file == nil {
		return nil, werror.WErrMsg("Could not find file to create lifetime with")
	}

	return &Lifetime{
		Id:         id,
		LiveFileId: createAction.GetDestinationId(),
		Actions:    []*FileAction{createAction.(*FileAction)},
		ContentId: file.GetContentId(),
		ServerId: InstanceService.GetLocal().ServerId(),
	}, nil
}

func (l *Lifetime) ID() LifetimeId {
	return l.Id
}

func (l *Lifetime) Add(action types.FileAction) {
	l.actionsLock.Lock()
	defer l.actionsLock.Unlock()

	action.SetLifetimeId(l.Id)
	l.Actions = append(l.Actions, action.(*FileAction))
	l.LiveFileId = action.GetDestinationId()
}

func (l *Lifetime) GetLatestFileId() FileId {
	return l.LiveFileId
}

func (l *Lifetime) GetLatestAction() types.FileAction {
	return l.Actions[len(l.Actions)-1]
}

func (l *Lifetime) GetContentId() weblens.ContentId {
	return l.ContentId
}

func (l *Lifetime) SetContentId(cId weblens.ContentId) {
	l.ContentId = cId
}

// IsLive returns a boolean representing if this Lifetime shows a file
// currently on the real filesystem, and has not been deleted.
func (l *Lifetime) IsLive() bool {
	// If the most recent action has no destination, the file was removed
	return l.Actions[len(l.Actions)-1].DestinationId != ""
}

func (l *Lifetime) GetActions() []types.FileAction {
	l.actionsLock.RLock()
	defer l.actionsLock.RUnlock()
	return internal.SliceConvert[types.FileAction](l.Actions)
}
