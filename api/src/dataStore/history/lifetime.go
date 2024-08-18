package history

import (
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Lifetime struct {
	Id         types.LifetimeId `bson:"_id" json:"id"`
	LiveFileId types.FileId     `bson:"liveFileId" json:"liveFileId"`
	ContentId  types.ContentId  `bson:"contentId,omitempty" json:"contentId,omitempty"`
	Actions    []*FileAction    `bson:"actions" json:"actions"`
	ServerId   types.InstanceId `bson:"serverId" json:"serverId"`

	actionsLock *sync.RWMutex
}

func NewLifetime(id types.LifetimeId, createAction types.FileAction) (types.Lifetime, error) {
	if createAction.GetActionType() != FileCreate {
		return nil, types.NewWeblensError("First Lifetime action must be of type FileCreate")
	}

	if id == "" {
		id = types.LifetimeId(primitive.NewObjectID().Hex())
	}

	createAction.SetLifetimeId(id)

	file := types.SERV.FileTree.Get(createAction.GetDestinationId())
	if file == nil {
		return nil, types.WeblensErrorMsg("Could not find file to create lifetime with")
	}

	return &Lifetime{
		Id:         id,
		LiveFileId: createAction.GetDestinationId(),
		Actions:    []*FileAction{createAction.(*FileAction)},
		ContentId: file.GetContentId(),
		ServerId:   types.SERV.InstanceService.GetLocal().ServerId(),

		actionsLock: &sync.RWMutex{},
	}, nil
}

func (l *Lifetime) ID() types.LifetimeId {
	return l.Id
}

func (l *Lifetime) Add(action types.FileAction) {
	if l.actionsLock == nil {
		l.actionsLock = &sync.RWMutex{}
	}
	l.actionsLock.Lock()
	defer l.actionsLock.Unlock()

	action.SetLifetimeId(l.Id)
	l.Actions = append(l.Actions, action.(*FileAction))
	l.LiveFileId = action.GetDestinationId()
}

func (l *Lifetime) GetLatestFileId() types.FileId {
	return l.LiveFileId
}

func (l *Lifetime) GetLatestAction() types.FileAction {
	return l.Actions[len(l.Actions)-1]
}

func (l *Lifetime) GetContentId() types.ContentId {
	return l.ContentId
}

func (l *Lifetime) SetContentId(cId types.ContentId) {
	l.ContentId = cId
}

// IsLive returns a boolean representing if this Lifetime shows a file
// currently on the real filesystem, and has not been deleted.
func (l *Lifetime) IsLive() bool {
	// If the most recent action has no destination, the file was removed
	return l.Actions[len(l.Actions)-1].DestinationId != ""
}

func (l *Lifetime) GetActions() []types.FileAction {
	if l.actionsLock == nil {
		l.actionsLock = &sync.RWMutex{}
	}
	l.actionsLock.RLock()
	defer l.actionsLock.RUnlock()
	return util.SliceConvert[types.FileAction](l.Actions)
}
