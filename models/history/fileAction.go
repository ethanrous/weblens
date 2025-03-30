package history

import (
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/context"
)

type FileActionType = string

const (
	FileCreate     FileActionType = "fileCreate"
	FileMove       FileActionType = "fileMove"
	FileSizeChange FileActionType = "fileSizeChange"
	Backup         FileActionType = "backup"
	FileDelete     FileActionType = "fileDelete"
	FileRestore    FileActionType = "fileRestore"
)

type FileAction struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	file            *file_model.WeblensFileImpl `bson:"-"`
	ActionType      FileActionType              `json:"actionType" bson:"actionType"`
	OriginPath      string                      `json:"originPath" bson:"originPath,omitempty"`
	DestinationPath string                      `json:"destinationPath" bson:"destinationPath,omitempty"`
	LifeId          string                      `json:"lifeId" bson:"lifeId"`
	EventId         FileEventId                 `json:"eventId" bson:"eventId"`
	ParentId        string                      `json:"parentId" bson:"parentId"`
	ServerId        string                      `json:"serverId" bson:"serverId"`

	Size int64 `json:"size" bson:"size"`
}

func (fa *FileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

func (fa *FileAction) SetSize(size int64) {
	fa.Size = size
}

func (fa *FileAction) GetSize() int64 {
	return fa.Size
}

func (fa *FileAction) GetFile() *file_model.WeblensFileImpl {
	return fa.file
}

func (fa *FileAction) SetLifetimeId(lId string) {
	fa.LifeId = lId
}

func (fa *FileAction) GetLifetimeId() string {
	return fa.LifeId
}

func (fa *FileAction) GetOriginPath() string {
	return fa.OriginPath
}

func (fa *FileAction) GetDestinationPath() string {
	return fa.DestinationPath
}

func (fa *FileAction) GetRelevantPath() string {
	if fa.GetActionType() == FileDelete {
		return fa.OriginPath
	}
	return fa.DestinationPath
}

func (fa *FileAction) GetActionType() FileActionType {
	return fa.ActionType
}

func (fa *FileAction) GetParentId() string {
	return fa.ParentId
}

const FileActionCollectionKey = "fileHistory"

func SaveAction(ctx context.DatabaseContext, action *FileAction) error {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, action)
	if err != nil {
		return err
	}

	return nil
}
