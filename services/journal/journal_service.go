package journal

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func NewEvent() *history.FileEvent {
	return &history.FileEvent{}
}

//	func GetActiveLifetimesByTowerId(ctx context.Context, towerId string) ([]*history.Lifetime, error) {
//		return nil, nil
//	}
func GetPastFile(ctx context.Context, id fs.Filepath, time time.Time) (*file_model.WeblensFileImpl, error) {
	// actions, err := getActionsByPath(ctx, id, false)
	// if err != nil {
	// 	return nil, err
	// }
	// slices.SortFunc(
	// 	actions, func(a, b *history.FileAction) int {
	// 		return a.GetTimestamp().Compare(b.GetTimestamp())
	// 	},
	// )
	//
	// // If the first action is after the time we are looking for, we need to get the actions
	// // from the path of the file, but not necessarily the same lifetime.
	// diff := actions[0].GetTimestamp().UnixMilli() - time.UnixMilli()
	// if time.Unix() != 0 && diff > 0 {
	// 	actions, err = j.getActionsByPath(lt.GetLatestPath(), true)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	slices.SortFunc(
	// 		actions, func(a, b *history.FileAction) int {
	// 			return a.GetTimestamp().Compare(b.GetTimestamp())
	// 		},
	// 	)
	// }
	//
	// if len(actions) == 0 {
	// 	return nil, errors.WithStack(file_model.ErrFileNotFound)
	// }
	//
	// relevantAction := actions[len(actions)-1]
	// counter := 1
	// for relevantAction.GetTimestamp().UnixMilli() >= time.UnixMilli() {
	// 	counter++
	// 	if len(actions)-counter < 0 {
	// 		break
	// 	}
	// 	if actions[len(actions)-counter].ActionType == history.FileSizeChange {
	// 		continue
	// 	}
	// 	relevantAction = actions[len(actions)-counter]
	// }
	//
	// if relevantAction.ActionType == history.FileDelete {
	// 	return nil, errors.Errorf("Trying to get past file after delete [%s]", id)
	// }
	//
	// if fs.IsZeroFilepath(relevantAction.DestinationPath) {
	// 	return nil, errors.Errorf("Got empty DestinationPath trying to get past file [%s] from journal", id)
	// }
	//
	// f := file_model.NewWeblensFile(relevantAction.DestinationPath)
	// // f.parentId = relevantAction.ParentId
	// // f.portablePath = path
	// // f.pastFile = true
	// // f.pastId = relevantAction.LifeId
	// // f.SetContentId(lt.ContentId)
	// // f.setModifyDate(relevantAction.GetTimestamp())
	//
	// children, err := GetPastFolderChildren(f, time)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// for _, child := range children {
	// 	err = f.AddChild(child)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	//
	// return f, nil
	return nil, errors.New("not implemented")
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

func GetActionsByPath(ctx context.Context, path fs.Filepath) ([]*history.FileAction, error) {
	return getActionsByPath(ctx, path, false)
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

func getActionsByPath(ctx context.Context, path fs.Filepath, noChildren bool) ([]*history.FileAction, error) {
	col, err := db.GetCollection(ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	var pathMatch bson.A
	if noChildren {
		pathMatch = bson.A{
			bson.D{{Key: "actions.originPath", Value: path.ToPortable()}},
			bson.D{{Key: "actions.destinationPath", Value: path.ToPortable()}},
		}
	} else {
		pathMatch = bson.A{
			bson.D{{Key: "actions.originPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
			bson.D{{Key: "actions.destinationPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
		}
	}

	pipe := bson.A{
		// bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathMatch}}}},
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
