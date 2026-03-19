// Package journal provides functionalities to manage and retrieve file action histories and lifetimes.
package journal

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	"go.mongodb.org/mongo-driver/bson"
)

// getPastFileIDAtPath retrieves the file ID that existed at the given path at a specific point in time.
// Returns the root alias for root paths, otherwise queries the journal to find the file ID.
func getPastFileIDAtPath(ctx context.Context, path wlfs.Filepath, time time.Time) (string, error) {
	if path.IsRoot() {
		// The root path's file ID is always the root alias
		return path.RootAlias, nil
	}

	actions, err := history.GetActionsAtPathBefore(ctx, path, time, history.GetActionsOptions{IncludeChildren: false})
	if err != nil {
		return "", err
	}

	if len(actions) != 1 {
		return "", wlerrors.Errorf("could not determine past file ID at path [%s] (ambiguous, %d actions found)", path, len(actions))
	}

	lastAction := actions[len(actions)-1]

	return lastAction.FileID, nil
}

// getPastFileChildren retrieves the children of a past file at a specific point in time.
func getPastFileChildren(ctx context.Context, pastFile *file_model.WeblensFileImpl, time time.Time) (map[wlfs.Filepath]*file_model.WeblensFileImpl, error) {
	if !pastFile.IsDir() {
		return nil, wlerrors.Wrapf(file_model.ErrDirectoryRequired, "cannot get children of non-directory file [%s]", pastFile.GetPortablePath())
	}

	path := pastFile.GetPortablePath()

	actions, err := history.GetActionsAtPathBefore(ctx, path, time, history.GetActionsOptions{IncludeChildren: true})
	if err != nil {
		return nil, err
	}

	childActions := make(map[wlfs.Filepath]history.FileAction)

	for _, action := range actions {
		if action.GetRelevantPath() == path {
			continue
		}

		pathKey := action.GetRelevantPath()
		if action.ActionType == history.FileMove && action.OriginPath.Dir() == path {
			pathKey = action.OriginPath
		}

		if _, ok := childActions[pathKey]; !ok {
			childActions[pathKey] = action
		}
	}

	children := make(map[wlfs.Filepath]*file_model.WeblensFileImpl)

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
// Unlike GetPastFileByPath, this uses the known file ID directly instead of looking it up by path,
// which avoids ambiguity when multiple files have existed at the same path over time.
func GetPastFileByID(ctx context.Context, fileID string, timestamp time.Time) (*file_model.WeblensFileImpl, error) {
	wlog.FromContext(ctx).Trace().Msgf("Getting past file by ID [%s] at time [%s]", fileID, timestamp)

	lastActionInPast, err := history.GetLastActionByFileIDBefore(ctx, fileID, timestamp)
	if err != nil && !db.IsNotFound(err) {
		return nil, err
	} else if db.IsNotFound(err) {
		// If we didn't find any actions for this file ID before the timestamp, it might be because the file was created after the timestamp.
		// We have to check if there was a file at this *path* in the past, at the timestamp, but perhaps with a different ID. This is a bit of an edge case,
		// but it can happen if a file was deleted and then a new file was created/moved/restored/etc at the same path.

		// Get the latest action of the current file ID to find its current path, and then check if there was a different file at that path in the past.
		currentFileLatestAction, err := history.GetLastActionByFileIDBefore(ctx, fileID, time.Now())
		if err != nil {
			return nil, err
		}

		viewingPath := currentFileLatestAction.GetRelevantPath()

		actions, err := history.GetActionsAtPathBefore(ctx, viewingPath, timestamp, history.GetActionsOptions{IncludeChildren: false})
		if err != nil {
			return nil, err
		}

		if len(actions) == 0 {
			return nil, wlerrors.Errorf("no actions found for path [%s] before time [%s]", viewingPath, timestamp)
		}

		// Choose the most recent action at that path before the timestamp, which should correspond to the file that existed at that path at that time.
		lastActionInPast = actions[len(actions)-1]
	}

	path := lastActionInPast.GetRelevantPath()

	newFile := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:         path,
		FileID:       fileID,
		IsPastFile:   true,
		ContentID:    lastActionInPast.ContentID,
		Size:         lastActionInPast.Size,
		ModifiedDate: option.Of(lastActionInPast.Timestamp),
	})

	// WARN: this might need to be uncommented
	if newFile.IsDir() {
		_, err = getPastFileChildren(ctx, newFile, timestamp)
		if err != nil {
			return nil, wlerrors.Errorf("failed to get past file children: %w", err)
		}
	}

	parentPath := path.Dir()

	parentFileID, err := getPastFileIDAtPath(ctx, parentPath, timestamp)
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

// GetPastFileByPath retrieves the historical state of a file at a given path and point in time.
// It reconstructs the file's state including its children and parent relationships as they existed at that time.
func GetPastFileByPath(ctx context.Context, path wlfs.Filepath, time time.Time) (*file_model.WeblensFileImpl, error) {
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

// GetActionsPage retrieves a paginated list of file actions.
func GetActionsPage(ctx context.Context, pageSize, pageNum int) ([]history.FileAction, error) {
	return history.GetActionsPage(ctx, pageSize, pageNum, "")
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
		return nil, wlerrors.WithStack(err)
	}

	var target []*history.FileAction

	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	return target, nil
}

// GetLatestPathByID retrieves the most recent path where a file with the given ID was located.
// Returns the destination path if available, otherwise falls back to the filepath field.
func GetLatestPathByID(ctx context.Context, fileID string) (wlfs.Filepath, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return wlfs.Filepath{}, err
	}

	pipe := bson.A{
		bson.M{"$match": bson.M{"fileID": fileID}},
		bson.M{"$sort": bson.M{"timestamp": -1}},
		bson.M{"$limit": 1},
		bson.M{"$project": bson.M{"destinationPath": 1, "filepath": 1}},
	}

	ret, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return wlfs.Filepath{}, wlerrors.WithStack(err)
	}

	var result struct {
		DestinationPath string `bson:"destinationPath"`
		Filepath        string `bson:"filepath"`
	}

	if !ret.Next(ctx) {
		return wlfs.Filepath{}, wlerrors.New("no results found")
	}

	err = ret.Decode(&result)
	if err != nil {
		return wlfs.Filepath{}, wlerrors.WithStack(err)
	}

	if result.DestinationPath != "" {
		return wlfs.ParsePortable(result.DestinationPath)
	}

	return wlfs.ParsePortable(result.Filepath)
}
