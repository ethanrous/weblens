package album

import (
	"fmt"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type Album struct {
	Id     types.AlbumId
	Name   string
	Owner  types.User
	Medias []types.Media
}

func New(albumName string, owner types.User) types.Album {
	albumId := types.AlbumId(util.GlobbyHash(12, fmt.Sprintln(albumName, owner.GetUsername())))
	return &Album{
		Id:     albumId,
		Name:   albumName,
		Owner:  owner,
		Medias: []types.Media{},
	}
}
