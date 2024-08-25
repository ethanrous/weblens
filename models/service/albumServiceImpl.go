package service

import (
	"context"
	"iter"
	"maps"
	"slices"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.AlbumService = (*AlbumServiceImpl)(nil)

type AlbumServiceImpl struct {
	albumsMap map[models.AlbumId]*models.Album

	mediaService *MediaServiceImpl
	collection   *mongo.Collection
}

func NewAlbumService(col *mongo.Collection, mediaService *MediaServiceImpl) *AlbumServiceImpl {
	return &AlbumServiceImpl{
		albumsMap: make(map[models.AlbumId]*models.Album),

		mediaService: mediaService,
		collection:   col,
	}
}

func (as *AlbumServiceImpl) Init() error {
	ret, err := as.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	var target = make([]*models.Album, 0)
	err = ret.All(context.Background(), &target)
	if err != nil {
		return err
	}

	for _, a := range target {
		as.albumsMap[a.ID()] = a
	}

	return nil
}

func (as *AlbumServiceImpl) GetAllByUser(u *models.User) []*models.Album {
	albs := slices.Collect(maps.Values(as.albumsMap))
	albs = internal.Filter(
		albs, func(t *models.Album) bool {
			return t.GetOwner() == u.GetUsername() || slices.Contains(t.GetSharedWith(), u.GetUsername())
		},
	)

	return albs
}

func (as *AlbumServiceImpl) Size() int {
	return len(as.albumsMap)
}

func (as *AlbumServiceImpl) Get(aId models.AlbumId) *models.Album {
	return as.albumsMap[aId]
}

func (as *AlbumServiceImpl) Add(a *models.Album) error {
	_, err := as.collection.InsertOne(context.Background(), a)
	if err != nil {
		return err
	}

	as.albumsMap[a.ID()] = a

	return nil
}

func (as *AlbumServiceImpl) Del(aId models.AlbumId) error {
	if _, ok := as.albumsMap[aId]; ok {
		filter := bson.M{"_id": aId}
		_, err := as.collection.DeleteOne(context.Background(), filter)
		if err != nil {
			return werror.WithStack(err)
		}

		delete(as.albumsMap, aId)
		return nil
	} else {
		return werror.ErrNoAlbum
	}
}

func (as *AlbumServiceImpl) RemoveMediaFromAny(mediaId models.ContentId) error {
	filter := bson.M{"medias": mediaId}
	ret, err := as.collection.Find(context.Background(), filter)
	if err != nil {
		return werror.WithStack(err)
	}

	var target = make([]*models.Album, 0)
	err = ret.All(context.Background(), &target)
	if err != nil {
		return werror.WithStack(err)
	}

	for _, album := range target {
		a := as.albumsMap[album.ID()]
		a.RemoveMedia(mediaId)
	}

	return nil
}

func (as *AlbumServiceImpl) SetAlbumCover(albumId models.AlbumId, cover *models.Media) error {
	album, ok := as.albumsMap[albumId]
	if !ok {
		return werror.ErrNoAlbum
	}

	colors, err := as.mediaService.GetProminentColors(cover)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": cover.ID(), "primaryColor": colors[0], "secondaryColor": colors[1]}}
	_, err = as.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return werror.WithStack(err)
	}

	album.SetCover(cover.ID(), colors[0], colors[1])

	return nil
}

func (as *AlbumServiceImpl) GetAlbumMedias(album *models.Album) iter.Seq[*models.Media] {
	return func(yeild func(*models.Media) bool) {
		for _, id := range album.Medias {
			m := as.mediaService.Get(id)
			if !yeild(m) {
				return
			}
		}
	}
}

func (as *AlbumServiceImpl) AddMediaToAlbum(album *models.Album, media ...*models.Media) error {
	if album == nil {
		return werror.ErrNoAlbum
	}

	mediaIds := internal.Map(
		media, func(m *models.Media) models.ContentId {
			return m.ID()
		},
	)

	filter := bson.M{"_id": album.ID()}
	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mediaIds}}}
	_, err := as.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return werror.WithStack(err)
	}

	album.Medias = append(album.Medias, mediaIds...)

	return nil
}

func (as *AlbumServiceImpl) RemoveMediaFromAlbum(album *models.Album, mediaIds ...models.ContentId) error {
	return werror.NotImplemented("RemoveMediaFromAlbum")
}

func (as *AlbumServiceImpl) AddUsersToAlbum(album *models.Album, us ...*models.User) error {
	werror.NotImplemented("AddUsersToAlbum, this should be done in the access service now")
	// filter := bson.M{"_id": album.ID()}
	// update := bson.M{"$addToSet": bson.M{}}
	// as.collection.UpdateOne()
	// err := types.SERV.StoreService.AddUsersToAlbum(a.ID(), us)
	// if err != nil {
	// 	return err
	// }
	//
	// a.SharedWith = internal.AddToSet(
	// 	a.SharedWith, internal.Map(
	// 		us, func(u *User) Username {
	// 			return u.GetUsername()
	// 		},
	// 	)...,
	// )

	return nil
}
