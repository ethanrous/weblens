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

func GetPastFileById(ctx context.Context, fileId string, time time.Time) (*file_model.WeblensFileImpl, error) {
	lastAction, err := history.GetLastActionByFileIdBefore(ctx, fileId, time)
	if err != nil {
		return nil, err
	}

	return GetPastFileByPath(ctx, lastAction.GetRelevantPath(), time)
}

func GetPastFileByPath(ctx context.Context, path fs.Filepath, time time.Time) (*file_model.WeblensFileImpl, error) {
	actions, err := history.GetActionsAtPathBefore(ctx, path, time, true)
	if err != nil {
		return nil, err
	}

	fileId := ""
	children := make(map[fs.Filepath]history.FileAction)

	for _, action := range actions {
		if action.GetRelevantPath() == path {
			fileId = action.FileId

			continue
		}

		pathKey := action.GetRelevantPath()
		if action.ActionType == history.FileMove && action.OriginPath.Dir() == path {
			pathKey = action.OriginPath
		}

		if _, ok := children[pathKey]; !ok {
			children[pathKey] = action
		}
	}

	newFile := file_model.NewWeblensFile(file_model.NewFileOptions{Path: path, FileId: fileId, IsPastFile: true})

	parentActions, err := history.GetActionsAtPathBefore(ctx, path.Dir(), time, false)
	if err != nil {
		return nil, err
	}

	lastParentAction := parentActions[len(parentActions)-1]
	parent := file_model.NewWeblensFile(file_model.NewFileOptions{Path: path.Dir(), FileId: lastParentAction.FileId})

	err = newFile.SetParent(parent)
	if err != nil {
		return nil, err
	}

	for pathKey, action := range children {
		destPath := action.GetRelevantPath()

		// If the destination is not the same as the path we are looking for, skip it
		if pathKey != destPath {
			continue
		}

		child := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       destPath,
			FileId:     action.FileId,
			IsPastFile: true,
			Size:       action.Size,
			ContentId:  action.ContentId,
		})

		err = child.SetParent(newFile)
		if err != nil {
			return nil, err
		}

		err = newFile.AddChild(child)
		if err != nil {
			return nil, err
		}
	}

	return newFile, nil
}

func GetPastFolderChildren(folder *file_model.WeblensFileImpl, time time.Time) (
	[]*file_model.WeblensFileImpl, error,
) {
	// var id = folder.ID()
	// if pastId := folder.GetPastId(); pastId != "" {
	// 	id = pastId
	// }
	//
	// actions, err := j.getChildrenAtTime(id, time)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// lifeIdMap := map[string]any{}
	// children := []*file_model.WeblensFileImpl{}
	// for _, action := range actions {
	// 	if action == nil {
	// 		continue
	// 	}
	// 	if _, ok := lifeIdMap[action.LifeId]; ok {
	// 		continue
	// 	}
	//
	// 	newChild := file_model.NewWeblensFile(
	// 		action.GetLifetimeId(), filepath.Base(action.DestinationPath), folder,
	// 		action.DestinationPath[len(action.DestinationPath)-1] == '/',
	// 	)
	// 	newChild.setModifyDate(time)
	// 	newChild.setPastFile(true)
	// 	newChild.size.Store(action.Size)
	// 	newChild.contentId = j.Get(action.LifeId).ContentId
	// 	children = append(
	// 		children, newChild,
	// 	)
	//
	// 	lifeIdMap[action.LifeId] = nil
	// }
	//
	// return children, nil
	return nil, errors.New("not implemented")
}

func GetActionsByPathSince(ctx context.Context, path fs.Filepath, since time.Time, noChildren bool) ([]history.FileAction, error) {
	col, err := db.GetCollection(ctx, history.FileHistoryCollectionKey)
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

func GetAllActionsByTowerId(ctx context.Context, towerId string) ([]*history.FileAction, error) {
	col, err := db.GetCollection(ctx, history.FileHistoryCollectionKey)
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
	col, err := db.GetCollection(ctx, history.FileHistoryCollectionKey)
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
	col, err := db.GetCollection(ctx, history.FileActionCollectionKey)
	if err != nil {
		return nil, err
	}

	o := compileLifetimeOpts(opts...)

	activeFilter := bson.M{}
	if o.ActiveOnly {
		activeFilter = bson.M{"$match": bson.M{"actionType": bson.M{"$ne": "fileDelete"}}}
	}

	pathFilter := bson.M{}
	if !o.PathPrefix.IsZero() {
		pathFilter = bson.M{"$match": pathPrefixReFilter(o.PathPrefix, o.Depth)}
	}

	pipe := bson.A{
		bson.M{"$match": bson.M{"towerId": towerId}},
		pathFilter,
		bson.M{
			"$group": bson.M{
				"_id":     "$fileId",
				"actions": bson.M{"$push": "$$ROOT"},
			},
		},
		activeFilter,
		bson.M{"$sort": bson.M{"timestamp": 1}},
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

// func getActionsByPath(ctx context.Context, path fs.Filepath, noChildren bool) ([]*history.FileAction, error) {
// 	col, err := db.GetCollection(ctx, history.FileHistoryCollectionKey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var pathMatch bson.A
// 	if noChildren {
// 		pathMatch = bson.A{
// 			bson.D{{Key: "actions.originPath", Value: path.ToPortable()}},
// 			bson.D{{Key: "actions.destinationPath", Value: path.ToPortable()}},
// 		}
// 	} else {
// 		pathMatch = bson.A{
// 			bson.D{{Key: "actions.originPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
// 			bson.D{{Key: "actions.destinationPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
// 		}
// 	}
//
// 	pipe := bson.A{
// 		// bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
// 		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathMatch}}}},
// 		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
// 		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
// 	}
//
// 	ret, err := col.Aggregate(context.Background(), pipe)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	var target []*history.FileAction
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	return target, nil
// }
