package fileTree

import (
	"time"
)

type FileActionType = string

type FileAction struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	file            *WeblensFileImpl `bson:"-"`
	ActionType      FileActionType   `json:"actionType" bson:"actionType"`
	OriginPath      string           `json:"originPath" bson:"originPath,omitempty"`
	DestinationPath string           `json:"destinationPath" bson:"destinationPath,omitempty"`
	LifeId          FileId           `json:"lifeId" bson:"lifeId"`
	EventId         FileEventId      `json:"eventId" bson:"eventId"`
	ParentId        FileId           `json:"parentId" bson:"parentId"`
	ServerId        string           `json:"serverId" bson:"serverId"`

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

func (fa *FileAction) GetFile() *WeblensFileImpl {
	return fa.file
}

func (fa *FileAction) SetLifetimeId(lId FileId) {
	fa.LifeId = lId
}

func (fa *FileAction) GetLifetimeId() FileId {
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

func (fa *FileAction) GetParentId() FileId {
	return fa.ParentId
}

// func (fa *FileAction) MarshalJSON() ([]byte, error) {
// 	data := map[string]any{
// 		"timestamp":       fa.Timestamp.UnixMilli(),
// 		"actionType":      fa.ActionType,
// 		"originPath":      fa.OriginPath,
// 		"destinationPath": fa.DestinationPath,
// 		"lifeId":          fa.LifeId,
// 		"eventId":         fa.EventId,
// 		"size":            fa.Size,
// 		"parentId":        fa.ParentId,
// 	}
//
// 	return json.Marshal(data)
// }
//
// func (fa *FileAction) UnmarshalJSON(bs []byte) error {
// 	data := map[string]any{}
// 	err := json.Unmarshal(bs, &data)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	fa.Timestamp = time.UnixMilli(int64(data["timestamp"].(float64)))
// 	fa.ActionType = FileActionType(data["actionType"].(string))
// 	fa.OriginPath = data["originPath"].(string)
// 	fa.DestinationPath = data["destinationPath"].(string)
// 	fa.LifeId = FileId(data["lifeId"].(string))
// 	fa.EventId = FileEventId(data["eventId"].(string))
// 	fa.Size = int64(data["size"].(float64))
// 	fa.ParentId = FileId(data["parentId"].(string))
//
// 	return nil
// }

const (
	FileCreate     FileActionType = "fileCreate"
	FileMove       FileActionType = "fileMove"
	FileSizeChange FileActionType = "fileSizeChange"
	Backup         FileActionType = "backup"
	FileDelete     FileActionType = "fileDelete"
	FileRestore    FileActionType = "fileRestore"
)
