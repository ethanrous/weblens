package history

import (
	"context"
	"regexp"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
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

const fileActionErrorLabel = "file action"

type FileAction struct {
	Id primitive.ObjectID `bson:"_id" json:"id"`

	Timestamp time.Time `bson:"timestamp" json:"timestamp"`

	ActionType      FileActionType `bson:"actionType" json:"actionType"`
	Filepath        fs.Filepath    `bson:"filepath,omitempty" json:"filepath"`
	OriginPath      fs.Filepath    `bson:"originPath,omitempty" json:"originPath"`
	DestinationPath fs.Filepath    `bson:"destinationPath,omitempty" json:"destinationPath"`
	EventId         string         `bson:"eventId" json:"eventId"`
	TowerId         string         `bson:"towerId" json:"towerId"`
	ContentId       string         `bson:"contentId,omitempty" json:"contentId"`
	FileId          string         `bson:"fileId" json:"fileId"`
	Doer            string         `bson:"doer" json:"doer"` // The user or system that performed the action

	Size int64 `bson:"size" json:"size"`

	file *file_model.WeblensFileImpl `bson:"-" json:"-"`
}

type FileLifetime struct {
	Id      string       `bson:"_id"`
	Actions []FileAction `bson:"actions"`
}

func NewCreateAction(ctx context.Context, file *file_model.WeblensFileImpl) FileAction {
	towerId := ctx.Value("towerId").(string)

	eventId := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventId = event.EventId
		eventTime = event.StartTime
	} else {
		eventId = primitive.NewObjectID().Hex()
	}

	fileId := file.ID()
	if fileId == "" {
		fileId = primitive.NewObjectID().Hex()
	}

	return FileAction{
		ActionType: FileCreate,
		ContentId:  file.GetContentId(),
		EventId:    eventId,
		FileId:     fileId,
		Filepath:   file.GetPortablePath(),
		Size:       file.Size(),
		Timestamp:  eventTime,
		TowerId:    towerId,
		Doer:       event.Doer,

		file: file,
	}
}

func NewMoveAction(ctx context.Context, originPath, destinationPath fs.Filepath, file *file_model.WeblensFileImpl) FileAction {
	// if destinationPath != file.GetPortablePath() {
	// 	panic(errors.New("destination path does not match file path"))
	// }

	towerId := ctx.Value("towerId").(string)

	eventId := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventId = event.EventId
		eventTime = event.StartTime
	} else {
		eventId = primitive.NewObjectID().Hex()
	}

	return FileAction{
		ActionType:      FileMove,
		DestinationPath: destinationPath,
		EventId:         eventId,
		FileId:          file.ID(),
		OriginPath:      originPath,
		Size:            file.Size(),
		Timestamp:       eventTime,
		TowerId:         towerId,
		ContentId:       file.GetContentId(),
		Doer:            event.Doer,
	}
}

func NewDeleteAction(ctx context.Context, file *file_model.WeblensFileImpl) FileAction {
	towerId := ctx.Value("towerId").(string)

	eventId := ""
	eventTime := time.Now()

	event, ok := FileEventFromContext(ctx)
	if ok {
		eventId = event.EventId
		eventTime = event.StartTime
	} else {
		eventId = primitive.NewObjectID().Hex()
	}

	return FileAction{
		ActionType: FileDelete,
		ContentId:  file.GetContentId(),
		EventId:    eventId,
		FileId:     file.ID(),
		Filepath:   file.GetPortablePath(),
		Size:       file.Size(),
		Timestamp:  eventTime,
		TowerId:    towerId,
		Doer:       event.Doer,
	}
}

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

func (fa *FileAction) SetFile(file *file_model.WeblensFileImpl) {
	fa.file = file
}

const FileActionCollectionKey = "fileHistory"

func missingContentId(ctx context.Context, action *FileAction) error {
	if action.ContentId == "" && action.Size > 0 && !action.GetRelevantPath().IsDir() {
		return errors.Errorf("action for [%s]s contentId is empty", action.GetRelevantPath())
	}

	return nil
}

// SaveAction saves a FileAction to the database.
// It returns an error if the operation fails.
func SaveAction(ctx context.Context, action *FileAction) error {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return err
	}

	if action.TowerId == "" {
		return errors.New("towerId is empty")
	}

	if err = missingContentId(ctx, action); err != nil {
		return err
	}

	if action.Id.IsZero() {
		action.Id = primitive.NewObjectID()
	}

	_, err = col.InsertOne(ctx, action)
	if err != nil {
		return err
	}

	return nil
}

func SaveActions(ctx context.Context, actions []FileAction) error {
	if len(actions) == 0 {
		return nil
	}

	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return err
	}

	for i, a := range actions {
		if err = missingContentId(ctx, &a); err != nil {
			return err
		}

		if a.Id.IsZero() {
			actions[i].Id = primitive.NewObjectID()
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

// GetLatestAction retrieves the latest FileAction from the database.
// It returns the latest FileAction and an error if the operation fails.
func GetLatestAction(ctx context_mod.DatabaseContext) (*FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	action := &FileAction{}
	filter := bson.M{}                                         // No filter to get all documents
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1}) // Sort by timestamp descending

	err = col.FindOne(ctx, filter, opts).Decode(action)
	if err != nil {
		return nil, db.WrapError(err, "failed to GetLatestAction")
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
func GetActionsByTowerId(ctx context_mod.DatabaseContext, towerId string) ([]*FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"towerId": towerId}
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

func GetActionAtFilepath(ctx context.Context, filepath fs.Filepath) (*FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
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

func GetLastActionByFileIdBefore(ctx context.Context, fileId string, ts time.Time) (action FileAction, err error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return
	}

	filter := bson.M{"fileId": fileId, "timestamp": bson.M{"$lte": ts}}

	err = col.FindOne(ctx, filter, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&action)
	if err != nil {
		return
	}

	return
}

func UpdateAction(ctx context.Context, action *FileAction) error {
	if action.Id.IsZero() {
		return errors.New("cannot update action with zero ID")
	}

	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": action.Id}

	_, err = col.ReplaceOne(ctx, filter, action)
	if err != nil {
		return db.WrapError(err, "UpdateAction")
	}

	return nil
}

func GetActionsAfter(ctx context.Context, timestamp time.Time) ([]FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"timestamp": bson.M{"$gt": timestamp}}

	cursor, err := col.Find(ctx, filter)
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

func GetActionsAtPathBefore(ctx context.Context, path fs.Filepath, timestamp time.Time, includeChildren bool) ([]FileAction, error) {
	col, err := db.GetCollection(ctx, FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	filter := pathPrefixReFilter(path)
	filter["timestamp"] = bson.M{"$lte": timestamp}

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

func pathPrefixReFilter(path fs.Filepath) bson.M {
	pathRe := regexp.QuoteMeta(path.ToPortable())
	pathRe += `[^/]*/?`

	return bson.M{
		"$or": bson.A{
			bson.M{"filepath": bson.M{"$regex": pathRe}},
			bson.M{"originPath": bson.M{"$regex": pathRe}},
			bson.M{"destinationPath": bson.M{"$regex": pathRe}},
		},
	}
}
