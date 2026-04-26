package zip

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	startup.RegisterHook(dropZips)
	startup.RegisterHook(registerIndexes)
}

var zipFileIDKey = "zipfileID_unique_index"
var indexModels = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "zipFileID", Value: 1}},
		Options: options.Index().SetUnique(true).SetName(zipFileIDKey),
	},
}

func registerIndexes(ctx context.Context, _ config.Provider) error {
	col, err := db.GetCollection[*Zip](ctx, ZipCollectionKey)
	if err != nil {
		return err
	}

	for _, idx := range indexModels {
		if err := col.NewIndex(idx); err != nil {
			return err
		}
	}

	return nil
}

func dropZips(ctx context.Context, _ config.Provider) error {
	col, err := db.GetCollection[*Zip](ctx, ZipCollectionKey)
	if err != nil {
		return err
	}

	err = col.Drop(ctx)
	if err != nil {
		return err
	}

	return nil
}
