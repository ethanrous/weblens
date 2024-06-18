package album

import (
	"fmt"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type Album struct {
	Id             types.AlbumId `bson:"_id"`
	Name           string        `bson:"name"`
	Owner          types.User    `bson:"owner"`
	Medias         []types.Media `bson:"medias"`
	Cover          types.Media   `bson:"cover"`
	PrimaryColor   string        `bson:"primaryColor"`
	SecondaryColor string        `bson:"secondaryColor"`
	SharedWith     []types.User  `bson:"sharedWith"`
	ShowOnTimeline bool          `bson:"showOnTimeline"`

	AlbumsManager *albumService
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

func (a *Album) ID() types.AlbumId {
	return a.Id
}

func (a *Album) AddMedia(ms ...types.Media) error {
	if ms == nil {
		return types.ErrNoMedia
	}

	a.Medias = append(a.Medias, ms...)
	err := a.AlbumsManager.db.AddMediaToAlbum(a.Id, util.Map(ms, func(m types.Media) types.ContentId {
		return m.ID()
	}))

	return err
}

func (a *Album) RemoveMedia(m ...types.ContentId) error {
	for _, m := range m {
		a.Medias, _, _ = util.YoinkFunc(a.Medias, func(f types.Media) bool {
			return f.ID() == m
		})
	}

	util.ShowErr(types.NewWeblensError("not impl"))

	return nil
}

func (a *Album) Rename(newName string) error {
	return types.NewWeblensError("not impl - album rename")
}

func (a *Album) SetCover(cover types.Media) error {
	a.Cover = cover

	return types.NewWeblensError("not impl - album set cover")
	return nil
}

func (a *Album) GetCover() types.Media {
	return a.Cover
}

func (a *Album) GetMedias() []types.Media {
	return a.Medias
}

func (a *Album) GetOwner() types.User {
	return a.Owner
}

func (a *Album) GetPrimaryColor() string {
	return a.PrimaryColor
}

func (a *Album) AddUsers(us ...types.User) error {
	return types.NewWeblensError("not impl - album add users")
}
func (a *Album) RemoveUsers(uns ...types.Username) error {
	return types.NewWeblensError("not impl - album remove users")
}

// func GetAlbumById(albumId types.AlbumId) (a *AlbumData, err error) {
// 	a, err = dbServer.GetAlbumById(albumId)
// 	return
// }
//
// func DeleteAlbum(albumId types.AlbumId) (err error) {
// 	err = dbServer.DeleteAlbum(albumId)
// 	return
// }
//
// func (a *AlbumData) CanUserAccess(username types.Username) bool {
// 	return a.Owner == username || slices.Contains(a.SharedWith, username)
// }
//
// func (a *AlbumData) AddMedia(ms []types.Media) (addedCount int, err error) {
// 	mIds := util.Map(ms, func(m types.Media) types.ContentId { return m.ID() })
// 	a.Medias = util.AddToSet(a.Medias, mIds)
// 	addedCount, err = dbServer.addMediaToAlbum(a.Id, mIds)
// 	return
// }
//
// func (a *AlbumData) Rename(newName string) (err error) {
// 	err = dbServer.setAlbumName(a.Id, newName)
// 	return
// }
//
// func (a *AlbumData) SetCover(mediaId types.ContentId, ft types.FileTree) error {
// 	m := a. media.MediaMapGet(mediaId)
// 	if m == nil {
// 		return ErrNoMedia
// 	}
// 	colors, err := m.(*media.media).getProminentColors(ft)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = dbServer.SetAlbumCover(a.Id, mediaId, colors[0], colors[1])
// 	if err != nil {
// 		return fmt.Errorf("failed to set album cover")
// 	}
// 	return nil
// }
//
// func (a *AlbumData) AddUsers(usernames []types.Username) (err error) {
// 	a.SharedWith = util.AddToSet(a.SharedWith, usernames)
// 	err = dbServer.shareAlbum(a.Id, usernames)
// 	return
// }
//
// func (a *AlbumData) RemoveUsers(usernames []types.Username) (err error) {
// 	a.SharedWith = util.Filter(a.SharedWith, func(s types.Username) bool { return !slices.Contains(usernames, s) })
// 	err = dbServer.unshareAlbum(a.Id, usernames)
// 	return
// }
