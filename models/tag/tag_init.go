// Package tag provides file tagging functionality for organizing files across folders.
package tag

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ownerNameIndexKey = "owner_name_unique_index"
const fileIDsIndexKey = "fileIDs_index"

// IndexModels defines MongoDB indexes for the tags collection.
var IndexModels = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "owner", Value: 1}, {Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetName(ownerNameIndexKey),
	},
	{
		Keys:    bson.D{{Key: "fileIDs", Value: 1}},
		Options: options.Index().SetName(fileIDsIndexKey),
	},
}

func init() {
	startup.RegisterHook(registerTagIndexes)
}

func registerTagIndexes(ctx context.Context, _ config.Provider) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	for _, idx := range IndexModels {
		if err := col.NewIndex(idx); err != nil {
			return err
		}
	}

	return nil
}
