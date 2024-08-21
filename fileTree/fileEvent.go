package fileTree

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileEventId string

type FileEvent struct {
	EventId     FileEventId   `bson:"_id"`
	Actions     []*FileAction `bson:"actions"`
	EventBegin  time.Time     `bson:"eventBegin"`
	ActionsLock sync.Mutex    `bson:"-"`
}

// NewFileEvent returns a FileEvent, a container for multiple FileActions that occur due to the
// same event (move, delete, etc.)
func NewFileEvent() *FileEvent {
	return &FileEvent{
		EventId: FileEventId(primitive.NewObjectID().Hex()),
		EventBegin:  time.Now(),
		Actions:     []*FileAction{},
		ActionsLock: sync.Mutex{},
	}
}

func (fe *FileEvent) GetEventId() FileEventId {
	return fe.EventId
}

func (fe *FileEvent) addAction(a *FileAction) {
	fe.ActionsLock.Lock()
	defer fe.ActionsLock.Unlock()

	fe.Actions = append(fe.Actions, a)
}

func (fe *FileEvent) GetActions() []*FileAction {
	return internal.SliceConvert[*FileAction](fe.Actions)
}

func (fe *FileEvent) NewCreateAction(file *WeblensFile) *FileAction {
	newAction := &FileAction{
		Timestamp:       time.Now(),
		ActionType:      FileCreate,
		DestinationPath: file.GetPortablePath().ToPortable(),
		DestinationId: file.ID(),
		EventId:         fe.EventId,
		ParentId:        file.GetParent().ID(),

		file: file,
	}

	fe.addAction(newAction)

	return newAction
}

func (fe *FileEvent) NewMoveAction(originId FileId, file *WeblensFile) *FileAction {
	lt := types.SERV.FileTree.GetJournal().GetLifetimeByFileId(originId)
	if lt == nil {
		wlog.Error.Println("Cannot not find existing lifetime for originId", originId)
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		wlog.Error.Println("File previous destination does not match move origin")
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

func (fe *FileEvent) NewDeleteAction(originId FileId) *FileAction {
	lt := types.SERV.FileTree.GetJournal().GetLifetimeByFileId(originId)
	if lt == nil {
		wlog.ShowErr(
			werror.WErrMsg(
				fmt.Sprintf(
					"Cannot not find existing lifetime for originId [%s]", originId,
				),
			),
		)
		return nil
	}
	latest := lt.GetLatestAction()

	if latest.GetDestinationId() != originId {
		wlog.Error.Println("File previous destination does not match move origin")
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
