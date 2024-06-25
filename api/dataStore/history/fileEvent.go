package history

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileEvent struct {
	EventId     types.FileEventId `bson:"_id"`
	Actions     []*FileAction     `bson:"Actions"`
	EventBegin  time.Time         `bson:"eventBegin"`
	ActionsLock *sync.Mutex       `bson:"-"`
}

// NewFileEvent returns a FileEvent, a container for multiple FileActions that occur due to the
// same event (move, delete, etc.)
func NewFileEvent() types.FileEvent {
	return &FileEvent{
		EventId:     types.FileEventId(primitive.NewObjectID().Hex()),
		EventBegin:  time.Now(),
		Actions:     []*FileAction{},
		ActionsLock: &sync.Mutex{},
	}
}

func (fe *FileEvent) GetEventId() types.FileEventId {
	return fe.EventId
}

func (fe *FileEvent) addAction(a types.FileAction) {
	fe.ActionsLock.Lock()
	defer fe.ActionsLock.Unlock()

	fe.Actions = append(fe.Actions, a.(*FileAction))
}

func (fe *FileEvent) GetActions() []types.FileAction {
	return util.SliceConvert[types.FileAction](fe.Actions)
}

func (fe *FileEvent) NewCreateAction(file types.WeblensFile) types.FileAction {
	newAction := &FileAction{
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: file.GetPortablePath().ToPortable(),
		DestinationId:   types.SERV.FileTree.GenerateFileId(file.GetAbsPath()),
		EventId:         fe.EventId,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewMoveAction(originId, destinationId types.FileId) types.FileAction {
	newAction := &FileAction{
		Timestamp:     time.Now(),
		ActionType:    FileCreate,
		OriginId:      originId,
		DestinationId: destinationId,
		EventId:       fe.EventId,
	}

	fe.addAction(newAction)

	return newAction
}

// func (fe *FileEvent) UnmarshalBSON(b []byte) error {
// 	target := &map[string]any{}
// 	err := bson.Unmarshal(b, target)
// 	if err != nil {
// 		return err
// 	}
// 	util.Debug.Println(target)
// 	return nil
// }

// func (fe *FileEvent) UnmarshalBSONValue(t bsontype.Type, value []byte) error {
// 	util.Debug.Println(t, value)
//
// 	return nil
// }
