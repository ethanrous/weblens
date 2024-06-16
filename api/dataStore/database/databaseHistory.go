package database

import (
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *databaseService) WriteFileEvent(fe types.FileEvent) error {
	_, err := db.fileHistory.InsertOne(db.ctx, fe)
	if err != nil {
		return err
	}

	return nil
}

func (db *databaseService) GetAllFileEvents(target []types.FileEvent) (*mongo.Cursor, error) {
	ret, err := db.fileHistory.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// err = ret.All(db.ctx, &target)
	// if err != nil {
	// 	return nil, err
	// }

	return ret, nil
}
