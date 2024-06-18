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

func (db *databaseService) AddMediaToAlbum(aId types.AlbumId, mIds []types.ContentId) error {
	filter := bson.M{"albumId": aId}
	update := bson.M{"$addToSet": bson.M{"medias": mIds}}
	_, err := db.albums.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) RemoveMediaFromAlbum(aId types.AlbumId, mId types.ContentId) error {
	panic("implement me")
}

func (db *databaseService) GetAlbumsByMedia(id types.ContentId) ([]types.Album, error) {
	filter := bson.M{"medias": id}
	ret, err := db.albums.Find(db.ctx, filter)
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
