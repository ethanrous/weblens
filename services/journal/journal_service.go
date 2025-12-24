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

type GetLifetimesOptions struct {
	ActiveOnly bool
	PathPrefix fs.Filepath
	Depth      int
}

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

func getPastFileIdAtPath(ctx context.Context, path fs.Filepath, time time.Time) (string, error) {
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

	return lastAction.FileId, nil
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
			FileId:       action.FileId,
			IsPastFile:   true,
			Size:         action.Size,
			ContentId:    action.ContentId,
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

func GetPastFileById(ctx context.Context, fileId string, time time.Time) (*file_model.WeblensFileImpl, error) {
	lastAction, err := history.GetLastActionByFileIdBefore(ctx, fileId, time)
	if err != nil {
		return nil, err
	}

	return GetPastFileByPath(ctx, lastAction.GetRelevantPath(), time)
}

func GetPastFileByPath(ctx context.Context, path fs.Filepath, time time.Time) (*file_model.WeblensFileImpl, error) {
	pastFileId, err := getPastFileIdAtPath(ctx, path, time)
	if err != nil {
		return nil, err
	}

	newFile := file_model.NewWeblensFile(file_model.NewFileOptions{Path: path, FileId: pastFileId, IsPastFile: true})

	_, err = getPastFileChildren(ctx, newFile, time)
	if err != nil {
		return nil, err
	}

	parentPath := path.Dir()

	parentFileId, err := getPastFileIdAtPath(ctx, parentPath, time)
	if err != nil {
		return nil, err
	}

	parent := file_model.NewWeblensFile(file_model.NewFileOptions{Path: parentPath, FileId: parentFileId, IsPastFile: true})

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

func GetActionsByPathSince(ctx context.Context, path fs.Filepath, since time.Time, noChildren bool) ([]history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	var pathMatch bson.M

	if noChildren {
		// pathMatch = bson.A{
		// 	bson.D{{Key: "actions.originPath", Value: path.ToPortable()}},
		// 	bson.D{{Key: "actions.destinationPath", Value: path.ToPortable()}},
		// }
	} else {
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

func GetActionsSince(ctx context.Context, time time.Time) ([]*history.FileAction, error) {
	return getActionsSince(ctx, time, "")
}

func GetActionsPage(ctx context.Context, pageSize, pageNum int) ([]history.FileAction, error) {
	return getActionsPage(ctx, pageSize, pageNum, "")
}

func GetAllActionsByTowerId(ctx context.Context, towerId string) ([]*history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	pipe := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: towerId}}}},
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

func GetLatestPathById(ctx context.Context, fileId string) (fs.Filepath, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return fs.Filepath{}, err
	}

	pipe := bson.A{
		bson.M{"$match": bson.M{"fileId": fileId}},
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
	} else {
		return fs.ParsePortable(result.Filepath)
	}
}

func GetLifetimesByTowerId(ctx context.Context, towerId string, opts ...GetLifetimesOptions) ([]history.FileLifetime, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	o := compileLifetimeOpts(opts...)

	pipe := bson.A{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "towerId", Value: towerId},
				{Key: "$or", Value: pathPrefixReFilter(o.PathPrefix, o.Depth)["$or"]},
			},
			},
		},
	}

	fileIdGroup := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$fileId"},
			{Key: "actions", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		},
		},
	}

	if o.ActiveOnly {
		pipe = append(pipe,
			bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: 1}}}},
			fileIdGroup,
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
					{Key: "originalGroupId", Value: "$_id"},
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
					{Key: "originalGroupId", Value: 0},
					{Key: "fileCreateAction", Value: 0},
					{Key: "fileCreateTimestamp", Value: 0},
					{Key: "fileCreateFilepath", Value: 0},
				},
				},
			},
		)
	} else {
		pipe = append(pipe, fileIdGroup)
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
