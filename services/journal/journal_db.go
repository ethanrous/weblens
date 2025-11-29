package journal

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func getActionsSince(ctx context.Context, date time.Time, serverId string) ([]*history.FileAction, error) {
	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	// TODO: Make this work for actions, not lifetimes
	pipe := bson.A{
		bson.D{
			{
				Key:   "$match",
				Value: bson.D{{Key: "actions.timestamp", Value: bson.D{{Key: "$gt", Value: date}}}, {Key: "serverId", Value: serverId}},
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
