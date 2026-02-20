// Package history provides functionality for tracking and managing file action history.
package history

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FileActionType represents the type of action performed on a file.
type FileActionType = string

// File action type constants define the different types of operations that can be performed on files.
const (
	FileCreate     FileActionType = "fileCreate"
	FileMove       FileActionType = "fileMove"
	FileSizeChange FileActionType = "fileSizeChange"
	Backup         FileActionType = "backup"
	FileDelete     FileActionType = "fileDelete"
	FileRestore    FileActionType = "fileRestore"
)

// FileAction represents a recorded action performed on a file in the system.
type FileAction struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`

	Timestamp time.Time `bson:"timestamp" json:"timestamp"`

	ActionType      FileActionType `bson:"actionType" json:"actionType"`
	Filepath        fs.Filepath    `bson:"filepath,omitempty" json:"filepath"`
	OriginPath      fs.Filepath    `bson:"originPath,omitempty" json:"originPath"`
	DestinationPath fs.Filepath    `bson:"destinationPath,omitempty" json:"destinationPath"`
	EventID         string         `bson:"eventID" json:"eventID"`
	TowerID         string         `bson:"towerID" json:"towerID"`
	ContentID       string         `bson:"contentID,omitempty" json:"contentID"`
	FileID          string         `bson:"fileID" json:"fileID"`
	Doer            string         `bson:"doer" json:"doer"` // The user or system that performed the action

	Size int64 `bson:"size" json:"size"`

	file *file_model.WeblensFileImpl `bson:"-" json:"-"`
}

// FileLifetime represents the complete history of actions performed on a file.
type FileLifetime struct {
	ID      string       `bson:"_id"`
	Actions []FileAction `bson:"actions"`
}

// NewCreateAction creates a new FileAction representing a file creation event.
func NewCreateAction(ctx context.Context, file *file_model.WeblensFileImpl) FileAction {
	towerID := ctx.Value("towerID").(string)

	eventID := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventID = event.EventID
		eventTime = event.StartTime
	} else {
		eventID = primitive.NewObjectID().Hex()
	}

	fileID := file.ID()
	if fileID == "" {
		fileID = primitive.NewObjectID().Hex()
	}

	if !file.IsDir() && file.GetContentID() == "" {
		err := wlerrors.Errorf("creating FileAction for file with empty content ID")
		log.FromContext(ctx).Warn().Stack().Err(err).Str("fileID", fileID).Str("filename", file.GetPortablePath().Filename()).Msg("Creating FileAction for file with empty content ID")
	}

	return FileAction{
		ActionType: FileCreate,
		ContentID:  file.GetContentID(),
		EventID:    eventID,
		FileID:     fileID,
		Filepath:   file.GetPortablePath(),
		Size:       file.Size(),
		Timestamp:  eventTime,
		TowerID:    towerID,
		Doer:       event.Doer,

		file: file,
	}
}

// NewMoveAction creates a new FileAction representing a file move event.
func NewMoveAction(ctx context.Context, originPath, destinationPath fs.Filepath, file *file_model.WeblensFileImpl) FileAction {
	towerID := ctx.Value("towerID").(string)

	eventID := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventID = event.EventID
		eventTime = event.StartTime
	} else {
		eventID = primitive.NewObjectID().Hex()
	}

	return FileAction{
		ActionType:      FileMove,
		DestinationPath: destinationPath,
		EventID:         eventID,
		FileID:          file.ID(),
		OriginPath:      originPath,
		Size:            file.Size(),
		Timestamp:       eventTime,
		TowerID:         towerID,
		ContentID:       file.GetContentID(),
		Doer:            event.Doer,
	}
}

// NewDeleteAction creates a new FileAction representing a file deletion event.
func NewDeleteAction(ctx context.Context, file *file_model.WeblensFileImpl) FileAction {
	towerID := ctx.Value("towerID").(string)

	eventID := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventID = event.EventID
		eventTime = event.StartTime
	} else {
		eventID = primitive.NewObjectID().Hex()
	}

	return FileAction{
		ActionType: FileDelete,
		ContentID:  file.GetContentID(),
		EventID:    eventID,
		FileID:     file.ID(),
		Filepath:   file.GetPortablePath(),
		Size:       file.Size(),
		Timestamp:  eventTime,
		TowerID:    towerID,
		Doer:       event.Doer,
	}
}

// MarshalBSON marshals the FileAction to BSON format for database storage.
func (fa *FileAction) MarshalBSON() ([]byte, error) {
	if fa.Size < 1 && fa.file != nil {
		fa.Size = fa.file.Size()
	}

	// Don't call this on the FileAction struct itself, as it will cause an infinite loop.
	type actionAlias FileAction

	return bson.Marshal((*actionAlias)(fa))
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

// GetOriginPath returns the origin path of the file action, or the Filepath if no OriginPath is present.
func (fa *FileAction) GetOriginPath() fs.Filepath {
	if fa.OriginPath.IsZero() {
		return fa.Filepath
	}

	return fa.OriginPath
}

// GetDestinationPath returns the destination path of the file action, or the Filepath if no DestinationPath is present.
func (fa *FileAction) GetDestinationPath() fs.Filepath {
	if fa.DestinationPath.IsZero() {
		return fa.Filepath
	}

	return fa.DestinationPath
}

// GetRelevantPath returns the relevant path of the file action.
// If the action type is FileDelete, it returns the origin path;
// otherwise, it returns the destination path.
func (fa *FileAction) GetRelevantPath() fs.Filepath {
	switch fa.ActionType {
	case FileCreate:
		return fa.Filepath
	case FileMove:
		return fa.DestinationPath
	case FileDelete:
		return fa.OriginPath
	default:
		return fs.Filepath{}
	}
}

// GetActionType returns the action type of the file action.
func (fa *FileAction) GetActionType() FileActionType {
	return fa.ActionType
}

// SetFile sets the associated file for this FileAction.
func (fa *FileAction) SetFile(file *file_model.WeblensFileImpl) {
	fa.file = file
}

// SaveAction saves a FileAction to the database.
// It returns an error if the operation fails.
func SaveAction(ctx context.Context, action *FileAction) error {
	if action.TowerID == "" {
		return wlerrors.New("TowerID is empty")
	}

	if action.FileID == "" {
		return wlerrors.New("FileID is empty")
	}

	col, err := db.GetCollection[*FileAction](ctx, FileHistoryCollectionKey)
	if err != nil {
		return err
	}

	if action.ID.IsZero() {
		action.ID = primitive.NewObjectID()
	}

	_, err = col.InsertOne(ctx, action)
	if err != nil {
		return err
	}

	return nil
}

// SaveActions saves multiple FileActions to the database in a single operation.
func SaveActions(ctx context.Context, actions []FileAction) error {
	if len(actions) == 0 {
		return nil
	}

	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return err
	}

	for i, a := range actions {
		if a.ID.IsZero() {
			actions[i].ID = primitive.NewObjectID()
		}
	}

	anyActions := make([]any, len(actions))
	for i, a := range actions {
		anyActions[i] = a
	}

	_, err = col.InsertMany(ctx, anyActions)
	if err != nil {
		return db.WrapError(err, "failed to SaveActions")
	}

	return nil
}

// ActionSorter sorts two FileActions based on their timestamps.
// If the timestamps are equal, it sorts by the path length.
func ActionSorter(a, b FileAction) int {
	timeDiff := a.GetTimestamp().Sub(b.GetTimestamp())
	if timeDiff != 0 {
		return int(timeDiff)
	}

	// If the creation time is the same, sort by the path length. This is to ensure parent directories are created before their children.
	return len(a.GetRelevantPath().RelPath) - len(b.GetRelevantPath().RelPath)
}

// GetLatestActionByTowerID retrieves the latest FileAction from the database belinging to a specific towerID.
// It returns the latest FileAction and an error if the operation fails.
func GetLatestActionByTowerID(ctx context_mod.DatabaseContext, towerID string) (*FileAction, error) {
	col, err := db.GetCollection[*FileAction](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"towerID": towerID}
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1}) // Sort by timestamp descending

	action := &FileAction{}

	err = col.FindOne(ctx, filter, opts).Decode(action)
	if err != nil {
		return nil, db.WrapError(err, "failed to GetLatestAction")
	}

	return action, nil
}

// GetActionsByTowerID retrieves all FileActions associated with a specific towerID from the database.
// It returns a slice of FileActions and an error if the operation fails.
//
// Parameters:
//   - ctx: The database context used for the operation.
//   - towerID: The ID of the tower for which to retrieve file actions.
//
// Returns:
//   - A slice of FileActions associated with the specified towerID.
//   - An error if the operation fails.
func GetActionsByTowerID(ctx context_mod.DatabaseContext, towerID string) ([]*FileAction, error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"towerID": towerID}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx) //nolint:errcheck

	var actions []*FileAction

	err = cursor.All(ctx, &actions)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

// GetActionAtFilepath retrieves the most recent FileAction for a given filepath.
func GetActionAtFilepath(ctx context.Context, filepath fs.Filepath) (*FileAction, error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"$or": bson.A{bson.M{"filepath": filepath.ToPortable()}, bson.M{"destinationPath": filepath.ToPortable()}}}
	action := &FileAction{}

	err = col.FindOne(ctx, filter, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(action)
	if err != nil {
		return nil, db.WrapError(err, "GetActionAtFilepath looking for %s", filepath)
	}

	return action, nil
}

// GetLastActionByFileIDBefore retrieves the most recent FileAction for a file before a given timestamp.
func GetLastActionByFileIDBefore(ctx context.Context, fileID string, ts time.Time) (action FileAction, err error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return
	}

	filter := bson.M{"fileID": fileID, "timestamp": bson.M{"$lte": ts}}

	err = col.FindOne(ctx, filter, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&action)
	if err != nil {
		return
	}

	return
}

// UpdateAction updates an existing FileAction in the database.
func UpdateAction(ctx context.Context, action *FileAction) error {
	if action.ID.IsZero() {
		return wlerrors.New("cannot update action with zero ID")
	}

	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": action.ID}

	_, err = col.ReplaceOne(ctx, filter, action)
	if err != nil {
		return db.WrapError(err, "UpdateAction")
	}

	return nil
}

// GetActionsAtPathBefore retrieves FileActions at a path before a given timestamp, optionally including child paths.
func GetActionsAtPathBefore(ctx context.Context, path fs.Filepath, timestamp time.Time, includeChildren bool) ([]FileAction, error) {
	return getActionsAtPath(ctx, path, timestamp, true, includeChildren)
}

// GetActionsAtPathAfter retrieves FileActions at a path after a given timestamp, optionally including child paths.
func GetActionsAtPathAfter(ctx context.Context, path fs.Filepath, timestamp time.Time, includeChildren bool) ([]FileAction, error) {
	return getActionsAtPath(ctx, path, timestamp, false, includeChildren)
}

func getActionsAtPath(ctx context.Context, path fs.Filepath, timestamp time.Time, before, includeChildren bool) ([]FileAction, error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{}

	if !path.IsZero() {
		depth := 0
		if includeChildren {
			depth = 1
		}

		filter = pathPrefixReFilter(path, depth)
	}

	if before {
		filter["timestamp"] = bson.M{"$lt": timestamp}
	} else {
		filter["timestamp"] = bson.M{"$gt": timestamp}
	}

	opts := options.Find().SetSort(bson.M{"timestamp": -1})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var actions []FileAction

	err = cursor.All(ctx, &actions)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

// GetActionsPage retrieves a paginated list of FileActions from the database.
func GetActionsPage(ctx context.Context, pageSize, pageNum int, _ string) ([]FileAction, error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	skip := pageSize * pageNum

	pipe := bson.A{
		// bson.D{
		// 	{
		// 		Key:   "$match",
		// 		Value: bson.D{{Key: "serverID", Value: serverID}},
		// 	},
		// },
		bson.D{{Key: "$sort", Value: bson.D{{Key: "actions.timestamp", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: pageSize}},
	}

	ret, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	var target []FileAction

	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	return target, nil
}

// pathPrefixReFilter creates a MongoDB query filter that matches file actions at the given path
// and its descendants up to the specified depth using regular expressions.
func pathPrefixReFilter(path fs.Filepath, depth int) bson.M {
	pathRe := regexp.QuoteMeta(path.ToPortable())
	pathRe += `([^/]+/?){0,` + strconv.Itoa(depth) + `}/?$`

	return bson.M{
		"$or": bson.A{
			bson.M{"filepath": bson.M{"$regex": pathRe}},
			bson.M{"originPath": bson.M{"$regex": pathRe}},
			bson.M{"destinationPath": bson.M{"$regex": pathRe}},
		},
	}
}
