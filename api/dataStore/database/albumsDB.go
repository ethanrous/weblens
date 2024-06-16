package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/album"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) GetAllAlbums() ([]types.Album, error) {
	ret, err := db.albums.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target = make([]*album.Album, 0)
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Album](target), nil
}
