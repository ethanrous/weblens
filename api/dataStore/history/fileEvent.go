package history

import (
	"context"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fileEvent struct {
	EventId     types.FileEventId `bson:"_id"`
	Actions     []*fileAction     `bson:"actions"`
	EventBegin  time.Time         `bson:"eventBegin"`
	ActionsLock *sync.Mutex       `bson:"-"`
}

// NewFileEvent returns a FileEvent, a container for multiple FileActions that occur due to the
// same event (move, delete, etc.)
func NewFileEvent() types.FileEvent {
	return &fileEvent{
		EventId:     types.FileEventId(primitive.NewObjectID().String()),
		EventBegin:  time.Now(),
		Actions:     []*fileAction{},
		ActionsLock: &sync.Mutex{},
	}
}

func GetAllFileEvents() ([]types.FileEvent, error) {
	target := make([]*fileEvent, 0)
	ret, err := hc.dbServer.GetAllFileEvents(nil)
	if err != nil {
		return nil, err
	}

	err = ret.All(context.TODO(), &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.FileEvent](target), nil
}

func (fe *fileEvent) GetEventId() types.FileEventId {
	return fe.EventId
}

func (fe *fileEvent) AddAction(a types.FileAction) {
	fe.ActionsLock.Lock()
	defer fe.ActionsLock.Unlock()

	fe.Actions = append(fe.Actions, a.(*fileAction))
}

func (fe *fileEvent) GetActions() []types.FileAction {
	return util.SliceConvert[types.FileAction](fe.Actions)
}

// func (fe *fileEvent) UnmarshalBSON(b []byte) error {
// 	target := &map[string]any{}
// 	err := bson.Unmarshal(b, target)
// 	if err != nil {
// 		return err
// 	}
// 	util.Debug.Println(target)
// 	return nil
// }

// func (fe *fileEvent) UnmarshalBSONValue(t bsontype.Type, value []byte) error {
// 	util.Debug.Println(t, value)
//
// 	return nil
// }
