package journal

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/option"
	"go.mongodb.org/mongo-driver/bson"
)

// GetLifetimesOptions specifies filtering options for retrieving file lifetimes.
type GetLifetimesOptions struct {
	ActiveOnly bool
	PathPrefix fs.Filepath
	Depth      int
}

// compileLifetimeOpts merges multiple GetLifetimesOptions into a single compiled option set.
// The function applies the last non-zero value for each option field, with a minimum depth of 1.
func compileLifetimeOpts(opts ...GetLifetimesOptions) GetLifetimesOptions {
	o := GetLifetimesOptions{}
	o.Depth = 1 // Minimum depth

	for _, opt := range opts {
		if !opt.PathPrefix.IsZero() {
			o.PathPrefix = opt.PathPrefix
		}

		if opt.ActiveOnly {
			o.ActiveOnly = opt.ActiveOnly
		}

		if opt.Depth != 0 && opt.Depth > o.Depth {
			o.Depth = opt.Depth
		}
	}

	return o
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

// getPastFileIDAtPath retrieves the file ID that existed at the given path at a specific point in time.
// Returns the root alias for root paths, otherwise queries the action history to find the file ID.
func getPastFileIDAtPath(ctx context.Context, path fs.Filepath, time time.Time) (string, error) {
	if path.IsRoot() {
		// The root path's file ID is always the root alias
		return path.RootAlias, nil
	}

	actions, err := history.GetActionsAtPathBefore(ctx, path, time, false)
	if err != nil {
		return "", err
	}

	if len(actions) != 1 {
		return "", errors.Errorf("could not determine past file ID at path [%s] (ambiguous, %d actions found)", path, len(actions))
	}

	lastAction := actions[len(actions)-1]

	return lastAction.FileID, nil
}

// getPastFileChildren retrieves the children of a past file at a specific point in time.
func getPastFileChildren(ctx context.Context, pastFile *file_model.WeblensFileImpl, time time.Time) (map[fs.Filepath]*file_model.WeblensFileImpl, error) {
	path := pastFile.GetPortablePath()

	actions, err := history.GetActionsAtPathBefore(ctx, path, time, true)
	if err != nil {
		return nil, err
	}

	childActions := make(map[fs.Filepath]history.FileAction)

	for _, action := range actions {
		if action.GetRelevantPath() == path {
			continue
		}

		log.FromContext(ctx).Debug().Msgf("Considering action for child: %s, parent path: %s", action.GetRelevantPath(), path)

		pathKey := action.GetRelevantPath()
		if action.ActionType == history.FileMove && action.OriginPath.Dir() == path {
			pathKey = action.OriginPath
		}

		if _, ok := childActions[pathKey]; !ok {
			childActions[pathKey] = action
		}
	}

	children := make(map[fs.Filepath]*file_model.WeblensFileImpl)

	for pathKey, action := range childActions {
		destPath := action.GetRelevantPath()

		// If the destination is not the same as the path we are looking for, skip it
		if pathKey != destPath {
			continue
		}

		child := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:         destPath,
			FileID:       action.FileID,
			IsPastFile:   true,
			Size:         action.Size,
			ContentID:    action.ContentID,
			ModifiedDate: option.Of(action.Timestamp),
		})

		err = child.SetParent(pastFile)
		if err != nil {
			return nil, err
		}

		err = pastFile.AddChild(child)
		if err != nil {
			return nil, err
		}

		children[destPath] = child
	}

	return children, nil
}

// GetPastFileByID retrieves the historical state of a file by its ID at a specific point in time.
// It finds the file's path at the given time and delegates to GetPastFileByPath.
func GetPastFileByID(ctx context.Context, fileID string, time time.Time) (*file_model.WeblensFileImpl, error) {
	lastAction, err := history.GetLastActionByFileIDBefore(ctx, fileID, time)
	if err != nil {
		return nil, err
	}

	return GetPastFileByPath(ctx, lastAction.GetRelevantPath(), time)
}

// GetPastFileByPath retrieves the historical state of a file at a given path and point in time.
// It reconstructs the file's state including its children and parent relationships as they existed at that time.
func GetPastFileByPath(ctx context.Context, path fs.Filepath, time time.Time) (*file_model.WeblensFileImpl, error) {
	pastFileID, err := getPastFileIDAtPath(ctx, path, time)
	if err != nil {
		return nil, err
	}

	newFile := file_model.NewWeblensFile(file_model.NewFileOptions{Path: path, FileID: pastFileID, IsPastFile: true})

	_, err = getPastFileChildren(ctx, newFile, time)
	if err != nil {
		return nil, err
	}

	parentPath := path.Dir()

	parentFileID, err := getPastFileIDAtPath(ctx, parentPath, time)
	if err != nil {
		return nil, err
	}

	parent := file_model.NewWeblensFile(file_model.NewFileOptions{Path: parentPath, FileID: parentFileID, IsPastFile: true})

	err = newFile.SetParent(parent)
	if err != nil {
		return nil, err
	}

	err = parent.AddChild(newFile)
	if err != nil {
		return nil, err
	}

	return newFile, nil
}

// GetActionsByPathSince retrieves all file actions at or under the given path since the specified time.
// The noChildren parameter controls whether to include descendant paths or only exact path matches.
func GetActionsByPathSince(ctx context.Context, path fs.Filepath, since time.Time, noChildren bool) ([]history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	var pathMatch bson.M

	if !noChildren {
		pathMatch = pathPrefixReFilter(path, 1)
	}

	pipe := bson.A{
		bson.M{"$match": bson.M{"timestamp": bson.M{"$gt": since}}},
		bson.M{"$match": pathMatch},
		// bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		// bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathMatch}}}},
		// bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
	}

	ret, err := col.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var target []history.FileAction

	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return target, nil
}

// GetActionsSince retrieves all file actions since the specified time.
func GetActionsSince(ctx context.Context, time time.Time) ([]*history.FileAction, error) {
	return getActionsSince(ctx, time, "")
}

// GetActionsPage retrieves a paginated list of file actions.
func GetActionsPage(ctx context.Context, pageSize, pageNum int) ([]history.FileAction, error) {
	return getActionsPage(ctx, pageSize, pageNum, "")
}

// GetAllActionsByTowerID retrieves all file actions associated with a specific tower,
// sorted by timestamp in descending order (most recent first).
func GetAllActionsByTowerID(ctx context.Context, towerID string) ([]*history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	pipe := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "serverID", Value: towerID}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
	}

	ret, err := col.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var target []*history.FileAction

	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return target, nil
}

// GetLatestPathByID retrieves the most recent path where a file with the given ID was located.
// Returns the destination path if available, otherwise falls back to the filepath field.
func GetLatestPathByID(ctx context.Context, fileID string) (fs.Filepath, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return fs.Filepath{}, err
	}

	pipe := bson.A{
		bson.M{"$match": bson.M{"fileID": fileID}},
		bson.M{"$sort": bson.M{"timestamp": -1}},
		bson.M{"$limit": 1},
		bson.M{"$project": bson.M{"destinationPath": 1, "filepath": 1}},
	}

	ret, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return fs.Filepath{}, errors.WithStack(err)
	}

	var result struct {
		DestinationPath string `bson:"destinationPath"`
		Filepath        string `bson:"filepath"`
	}

	if !ret.Next(ctx) {
		return fs.Filepath{}, errors.New("no results found")
	}

	err = ret.Decode(&result)
	if err != nil {
		return fs.Filepath{}, errors.WithStack(err)
	}

	if result.DestinationPath != "" {
		return fs.ParsePortable(result.DestinationPath)
	}

	return fs.ParsePortable(result.Filepath)
}

// GetLifetimesByTowerID retrieves file lifetimes (grouped actions by file ID) for files on a specific tower.
// The opts parameter allows filtering by path prefix, depth, and whether to include only active (non-deleted) files.
func GetLifetimesByTowerID(ctx context.Context, towerID string, opts ...GetLifetimesOptions) ([]history.FileLifetime, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	o := compileLifetimeOpts(opts...)

	pipe := bson.A{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "towerID", Value: towerID},
				{Key: "$or", Value: pathPrefixReFilter(o.PathPrefix, o.Depth)["$or"]},
			},
			},
		},
	}

	fileIDGroup := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$fileID"},
			{Key: "actions", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		},
		},
	}

	if o.ActiveOnly {
		pipe = append(pipe,
			bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: 1}}}},
			fileIDGroup,
			bson.D{{Key: "$match", Value: bson.D{{Key: "actions", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "actionType", Value: "fileDelete"}}}}}}}}}},
			bson.D{
				{Key: "$addFields", Value: bson.D{
					{Key: "fileCreateAction", Value: bson.D{
						{Key: "$first", Value: bson.D{
							{Key: "$filter", Value: bson.D{
								{Key: "input", Value: "$actions"},
								{Key: "as", Value: "a"},
								{Key: "cond", Value: bson.D{
									{Key: "$eq", Value: bson.A{
										"$$a.actionType",
										"fileCreate",
									},
									},
								},
								},
							},
							},
						},
						},
					},
					},
				},
				},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "originalGroupID", Value: "$_id"},
					{Key: "actions", Value: 1},
					{Key: "fileCreateAction", Value: 1},
					{Key: "fileCreateTimestamp", Value: "$fileCreateAction.timestamp"},
					{Key: "fileCreateFilepath", Value: "$fileCreateAction.filepath"},
				},
				},
			},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "fileCreateAction.timestamp", Value: -1}}}},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$fileCreateAction.filepath"},
					{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
				},
				},
			},
			bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$doc"}}}},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "originalGroupID", Value: 0},
					{Key: "fileCreateAction", Value: 0},
					{Key: "fileCreateTimestamp", Value: 0},
					{Key: "fileCreateFilepath", Value: 0},
				},
				},
			},
		)
	} else {
		pipe = append(pipe, fileIDGroup)
	}

	cur, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	var lifetimes []history.FileLifetime

	err = cur.All(ctx, &lifetimes)
	if err != nil {
		return nil, db.WrapError(err, "GetLifetimes")
	}

	return lifetimes, nil
}
