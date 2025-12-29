// Package journal handles database operations related to journaling file actions.
package journal

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func getActionsSince(ctx context.Context, date time.Time, serverID string) ([]*history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	pipe := bson.A{
		bson.D{
			{
				Key:   "$match",
				Value: bson.D{{Key: "actions.timestamp", Value: bson.D{{Key: "$gt", Value: date}}}, {Key: "serverID", Value: serverID}},
			},
		},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "actions.timestamp", Value: 1}}}},
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

func getActionsPage(ctx context.Context, pageSize, pageNum int, _ string) ([]history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
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
		return nil, errors.WithStack(err)
	}

	var target []history.FileAction

	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return target, nil
}
