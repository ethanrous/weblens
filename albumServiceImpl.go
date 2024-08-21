package weblens

import (
	"context"
	"maps"
	"slices"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlbumServiceImpl struct {
	albumsMap  map[AlbumId]*Album
	collection *mongo.Collection
}

func NewAlbumService(col *mongo.Collection) *AlbumServiceImpl {
	return &AlbumServiceImpl{
		albumsMap:  make(map[AlbumId]*Album),
		collection: col,
	}
}

func (as *AlbumServiceImpl) Init() error {
	ret, err := as.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	var target = make([]*Album, 0)
	err = ret.All(context.Background(), &target)
	if err != nil {
		return err
	}

	for _, a := range target {
		as.albumsMap[a.ID()] = a
	}

	return nil
}

func (as *AlbumServiceImpl) GetAllByUser(u types.User) []Album {
	albs := slices.Collect(maps.Values(as.albumsMap))
	albs = internal.Filter(
		albs, func(t *Album) bool {
			return t.GetOwner() == u || slices.Contains(t.GetSharedWith(), u.GetUsername())
		},
	)

	return internal.SliceConvert[Album](albs)
}

func (as *AlbumServiceImpl) Size() int {
	return len(as.albumsMap)
}

func (as *AlbumServiceImpl) Get(aId AlbumId) *Album {
	return as.albumsMap[aId]
}

func (as *AlbumServiceImpl) Add(a *Album) error {
	_, err := as.collection.InsertOne(context.Background(), a)
	if err != nil {
		return err
	}

	as.albumsMap[a.ID()] = a

	return nil
}

func (as *AlbumServiceImpl) Del(aId AlbumId) error {
	if _, ok := as.albumsMap[aId]; ok {
		filter := bson.M{"_id": aId}
		_, err := as.collection.DeleteOne(context.Background(), filter)
		if err != nil {
			return werror.Wrap(err)
		}

		delete(as.albumsMap, aId)
		return nil
	} else {
		return werror.ErrNoAlbum
	}
}

func (as *AlbumServiceImpl) RemoveMediaFromAny(mediaId ContentId) werror.WErr {
	filter := bson.M{"medias": mediaId}
	ret, err := as.collection.Find(context.Background(), filter)
	if err != nil {
		return werror.Wrap(err)
	}

	var target = make([]*Album, 0)
	err = ret.All(context.Background(), &target)
	if err != nil {
		return werror.Wrap(err)
	}

	for _, album := range target {
		a := as.albumsMap[album.ID()]
		werr := a.RemoveMedia(mediaId)
		if werr != nil {
			return werr
		}
	}

	return nil
}

func (as *AlbumServiceImpl) SetAlbumCover(albumId AlbumId, cover types.Media) werror.WErr {
	album, ok := as.albumsMap[albumId]
	if !ok {
		return werror.ErrNoAlbum
	}

	colors, werr := cover.GetProminentColors()
	if werr != nil {
		return werr
	}

	filter := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": cover.ID(), "primaryColor": colors[0], "secondaryColor": colors[1]}}
	_, err := as.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return werror.Wrap(err)
	}

	album.setCover(cover.ID(), colors[0], colors[1])

	return nil
}

func (as *AlbumServiceImpl) AddMediaToAlbum(aId AlbumId, media ...*Media) werror.WErr {
	album := as.albumsMap[aId]
	if album == nil {
		return werror.ErrNoAlbum
	}

	mediaIds := internal.Map(
		media, func(m *Media) ContentId {
			return m.ID()
		},
	)

	filter := bson.M{"_id": aId}
	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mediaIds}}}
	_, err := as.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return werror.Wrap(err)
	}

	album.Medias = append(album.Medias, mediaIds...)

	return nil
}
