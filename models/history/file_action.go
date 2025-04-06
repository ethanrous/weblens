package history

import (
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/fs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

var ErrFileActionNotFound = db.NewNotFoundError("file action")

type FileAction struct {
	Id primitive.ObjectID `json:"id" bson:"_id"`

	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	ActionType      FileActionType `json:"actionType" bson:"actionType"`
	Filepath        fs.Filepath    `json:"filepath" bson:"filepath,omitempty"`
	OriginPath      fs.Filepath    `json:"originPath" bson:"originPath,omitempty"`
	DestinationPath fs.Filepath    `json:"destinationPath" bson:"destinationPath,omitempty"`
	EventId         string         `json:"eventId" bson:"eventId"`
	TowerId         string         `json:"serverId" bson:"serverId"`
	ContentId       string         `json:"contentId" bson:"contentId"`

	Size int64 `json:"size" bson:"size"`
}

// GetTimestamp returns the timestamp of the file action.
func (fa *FileAction) GetTimestamp() time.Time {
	return fa.Timestamp
}

// SetSize sets the size of the file action.
func (fa *FileAction) SetSize(size int64) {
	fa.Size = size
}

// GetSize returns the size of the file action.
func (fa *FileAction) GetSize() int64 {
	return fa.Size
}

// GetOriginPath returns the origin path of the file action.
func (fa *FileAction) GetOriginPath() fs.Filepath {
	return fa.OriginPath
}

// GetDestinationPath returns the destination path of the file action.
func (fa *FileAction) GetDestinationPath() fs.Filepath {
	return fa.DestinationPath
}

// GetRelevantPath returns the relevant path of the file action.
// If the action type is FileDelete, it returns the origin path;
// otherwise, it returns the destination path.
func (fa *FileAction) GetRelevantPath() fs.Filepath {
	if fa.GetActionType() == FileDelete {
		return fa.OriginPath
	}
	return fa.DestinationPath
}

// GetActionType returns the action type of the file action.
func (fa *FileAction) GetActionType() FileActionType {
	return fa.ActionType
}

const FileActionCollectionKey = "fileHistory"

// SaveAction saves a FileAction to the database.
// It returns an error if the operation fails.
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

// ActionSorter sorts two FileActions based on their timestamps.
// If the timestamps are equal, it sorts by the path length.
func ActionSorter(a, b *FileAction) int {
	timeDiff := a.GetTimestamp().Sub(b.GetTimestamp())
	if timeDiff != 0 {
		return int(timeDiff)
	}

	// If the creation time is the same, sort by the path length. This is to ensure parent directories are created before their children.
	return len(a.DestinationPath.RelPath) - len(b.DestinationPath.RelPath)
}

// SaveFileAction saves a FileAction to the database.
// It returns an error if the operation fails.
func SaveFileAction(ctx context.DatabaseContext, action *FileAction) error {
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

// GetLatestAction retrieves the latest FileAction from the database.
// It returns the latest FileAction and an error if the operation fails.
func GetLatestAction(ctx context.DatabaseContext) (*FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	action := &FileAction{}
	filter := bson.M{}                                         // No filter to get all documents
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1}) // Sort by timestamp descending

	err = col.FindOne(ctx, filter, opts).Decode(action)
	if err != nil {
		return nil, err
	}

	return action, nil
}

// GetActionsByTowerId retrieves all FileActions associated with a specific towerId from the database.
// It returns a slice of FileActions and an error if the operation fails.
//
// Parameters:
//   - ctx: The database context used for the operation.
//   - towerId: The ID of the tower for which to retrieve file actions.
//
// Returns:
//   - A slice of FileActions associated with the specified towerId.
//   - An error if the operation fails.
func GetActionsByTowerId(ctx context.DatabaseContext, towerId string) ([]*FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"serverId": towerId}
	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var actions []*FileAction
	err = cursor.All(ctx, &actions)
	if err != nil {
		return nil, err
	}

	return actions, nil
}
