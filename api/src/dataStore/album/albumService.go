package album

import (
	"slices"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type albumService struct {
	repo map[types.AlbumId]*Album
	db   types.AlbumsStore
}

func NewService() types.AlbumService {
	return &albumService{
		repo: make(map[types.AlbumId]*Album),
	}
}

func (as *albumService) Init(db types.AlbumsStore) error {
	as.db = db
	albs, err := db.GetAllAlbums()
	if err != nil {
		return err
	}

	for _, a := range albs {
		as.repo[a.ID()] = a.(*Album)
	}

	return nil
}

func (as *albumService) GetAllByUser(u types.User) []types.Album {
	albs := util.MapToValues(as.repo)
	albs = util.Filter(
		albs, func(t *Album) bool {
			return t.GetOwner() == u || slices.Contains(t.GetSharedWith(), u.GetUsername())
		},
	)

	return util.SliceConvert[types.Album](albs)
}

func (as *albumService) Size() int {
	return len(as.repo)
}

func (as *albumService) Get(aId types.AlbumId) types.Album {
	return as.repo[aId]
}

func (as *albumService) Add(a types.Album) error {
	err := as.db.CreateAlbum(a)
	if err != nil {
		return err
	}

	as.repo[a.ID()] = a.(*Album)

	return nil
}

func (as *albumService) Del(aId types.AlbumId) error {
	if _, ok := as.repo[aId]; ok {
		err := as.db.DeleteAlbum(aId)
		if err != nil {
			return err
		}
		delete(as.repo, aId)
		return nil
	} else {
		return ErrNoAlbum
	}
}

func (as *albumService) RemoveMediaFromAny(mediaId types.ContentId) error {
	albums, err := as.db.GetAlbumsByMedia(mediaId)
	if err != nil {
		return err
	}

	for _, album := range albums {
		a := as.repo[album.ID()]
		err = a.RemoveMedia(mediaId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (as *albumService) SetAlbumCover(albumId types.AlbumId, cover types.Media) error {
	album, ok := as.repo[albumId]
	if !ok {
		return ErrNoAlbum
	}

	colors, err := cover.GetProminentColors()
	if err != nil {
		return err
	}

	err = as.db.SetAlbumCover(albumId, colors[0], colors[1], cover.ID())
	if err != nil {
		return err
	}

	album.setCover(cover.ID(), colors[0], colors[1])

	return nil
}

// Error

var ErrNoAlbum = types.NewWeblensError("album not found")
