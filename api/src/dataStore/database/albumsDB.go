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

func (db *databaseService) CreateAlbum(a types.Album) error {
	_, err := db.albums.InsertOne(db.ctx, a)
	return err
}

func (db *databaseService) AddMediaToAlbum(aId types.AlbumId, mIds []types.ContentId) error {
	filter := bson.M{"_id": aId}
	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mIds}}}
	_, err := db.albums.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) SetAlbumCover(aId types.AlbumId, color1, color2 string, mId types.ContentId) error {
	filter := bson.M{"_id": aId}
	update := bson.M{"$set": bson.M{"cover": mId, "primaryColor": color1, "secondaryColor": color2}}
	_, err := db.albums.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) RemoveMediaFromAlbum(aId types.AlbumId, mId types.ContentId) error {
	return types.ErrNotImplemented("removeMediaFromAlbum not implemented")
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

func (db *databaseService) AddUsersToAlbum(aId types.AlbumId, us []types.User) error {
	filter := bson.M{"_id": aId}
	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": util.Map(us, func(u types.User) types.Username {
		return u.GetUsername()
	})}}}

	_, err := db.albums.UpdateOne(db.ctx, filter, update)
	if err != nil {
		return err
	}

	return nil

}

func (db *databaseService) DeleteAlbum(id types.AlbumId) error {
	filter := bson.M{"_id": id}
	_, err := db.albums.DeleteOne(db.ctx, filter)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}
	return nil
}