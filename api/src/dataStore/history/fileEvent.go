package history

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileEvent struct {
	EventId     types.FileEventId `bson:"_id"`
	Actions     []*FileAction     `bson:"actions"`
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
		ParentId:        file.GetParent().ID(),

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewMoveAction(originId types.FileId, file types.WeblensFile) types.FileAction {
	lt := types.SERV.FileTree.GetJournal().GetLifetimeByFileId(originId)
	if lt == nil {
		util.Error.Println("Cannot not find existing lifetime for originId", originId)
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		util.Error.Println("File previous destination does not match move origin")
	}

	newAction := &FileAction{
		Timestamp:       time.Now(),
		ActionType:      FileMove,
		OriginId:        latest.GetDestinationId(),
		OriginPath:      latest.GetDestinationPath(),
		DestinationId:   file.ID(),
		DestinationPath: file.GetPortablePath().ToPortable(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewDeleteAction(originId types.FileId) types.FileAction {
	lt := types.SERV.FileTree.GetJournal().GetLifetimeByFileId(originId)
	if lt == nil {
		util.ShowErr(
			types.WeblensErrorMsg(
				fmt.Sprintf(
					"Cannot not find existing lifetime for originId [%s]", originId,
				),
			),
		)
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		util.Error.Println("File previous destination does not match move origin")
	}

	newAction := &FileAction{
		Timestamp:  time.Now(),
		ActionType: FileDelete,
		OriginId:   latest.GetDestinationId(),
		OriginPath: latest.GetDestinationPath(),
		EventId:    fe.EventId,
	}

	fe.addAction(newAction)

	return newAction
}
