package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) GetAllMedia() ([]types.Media, error) {
	ret, err := db.media.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target = make([]media.Media, 0)
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Media](target), nil
}
