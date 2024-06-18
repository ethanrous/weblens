package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) WriteFileEvent(fe types.FileEvent) error {
	_, err := db.fileHistory.InsertOne(db.ctx, fe)
	if err != nil {
		return err
	}

	return nil
}

func (db *databaseService) GetAllLifetimes() ([][]types.FileAction, error) {
	ret, err := db.fileHistory.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target [][]*history.FileEvent
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[[]types.FileAction](target), nil
}
