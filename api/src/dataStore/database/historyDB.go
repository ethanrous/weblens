package database

import (
	"fmt"

	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *databaseService) WriteFileEvent(fe types.FileEvent) error {
	_, err := db.fileHistory.InsertOne(db.ctx, fe)
	if err != nil {
		return err
	}

	return nil
}

func (db *databaseService) GetAllLifetimes() ([]types.Lifetime, error) {
	ret, err := db.fileHistory.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target []*history.Lifetime
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Lifetime](target), nil
}

func (db *databaseService) AddOrUpdateLifetime(lt types.Lifetime) error {
	filter := bson.M{"_id": lt.ID()}
	update := bson.M{"$set": lt}
	o := options.Update().SetUpsert(true)
	_, err := db.fileHistory.UpdateOne(db.ctx, filter, update, o)

	return err
}

func (db *databaseService) GetActionsByPath(path types.WeblensFilepath) ([]types.FileAction, error) {
	pipe := bson.A{
		bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
		bson.D{
			{
				"$match",
				bson.D{{"actions.destinationpath", bson.D{{"$regex", fmt.Sprint(path.ToPortable(), "[^/]*/?$")}}}},
			},
		},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
		bson.D{{"$sort", bson.D{{"timestamp", -1}}}},
	}

	ret, err := db.fileHistory.Aggregate(db.ctx, pipe)
	if err != nil {
		return nil, err
	}

	var target []*history.FileAction
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.FileAction](target), nil
}

func (db *databaseService) DeleteAllFileHistory() error {
	_, err := db.fileHistory.DeleteMany(db.ctx, bson.M{})
	return err
}