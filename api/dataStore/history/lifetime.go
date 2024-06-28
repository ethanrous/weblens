package history

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Lifetime struct {
	Id         types.LifetimeId `bson:"_id"`
	LiveFileId types.FileId
	ContentId  types.ContentId
	Actions    []*FileAction
}

func NewLifetime(id types.LifetimeId, createAction types.FileAction) (types.Lifetime, error) {
	if createAction.GetActionType() != FileCreate {
		return nil, types.NewWeblensError("First Lifetime action must be of type FileCreate")
	}

	if id == "" {
		id = types.LifetimeId(primitive.NewObjectID().Hex())
	}

	return &Lifetime{
		Id:         id,
		LiveFileId: createAction.GetDestinationId(),
		Actions:    []*FileAction{createAction.(*FileAction)},
		ContentId:  types.SERV.FileTree.Get(createAction.GetDestinationId()).GetContentId(),
	}, nil
}

func (l *Lifetime) ID() types.LifetimeId {
	return l.Id
}

func (l *Lifetime) Add(action types.FileAction) {
	action.SetLifetimeId(l.Id)
	l.Actions = append(l.Actions, action.(*FileAction))
	l.LiveFileId = action.GetDestinationId()
}

func (l *Lifetime) GetLatestFileId() types.FileId {
	return l.LiveFileId
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
	return util.SliceConvert[types.FileAction](l.Actions)
}
