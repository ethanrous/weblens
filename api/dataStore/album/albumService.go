package album

import (
	"github.com/ethrousseau/weblens/api/types"
)

type albumService struct {
	repo map[types.AlbumId]types.Album
	db   types.AlbumsDB
}

func NewService(db types.AlbumsDB) types.AlbumService {
	return &albumService{
		repo: make(map[types.AlbumId]types.Album),
		db:   db,
	}
}

func (as *albumService) Add(a types.Album) error {
	as.repo[a.ID()] = a
	return nil
}
