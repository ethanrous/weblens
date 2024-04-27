package dataStore

import (
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func CreateAlbum(albumName string, owner types.User) error {
	albumId := types.AlbumId(util.GlobbyHash(12, fmt.Sprintln(albumName, owner.GetUsername())))
	a := AlbumData{
		Id: albumId, Name: albumName,
		Owner:          owner.GetUsername(),
		ShowOnTimeline: true,
		Medias:         []types.ContentId{},
		SharedWith:     []types.Username{},
	}

	err := fddb.insertAlbum(a)
	if err != nil {
		return err
	}

	return nil
}

func VerifyAlbumsMedia() {
	as := fddb.getAllAlbums()
	if len(as) == 0 {
		return
	}

	for _, a := range as {
		deadMs := util.Filter(a.Medias, func(m types.ContentId) bool { return MediaMapGet(m) == nil })
		err := fddb.removeMediaFromAlbum(a.Id, deadMs)
		if err != nil {
			panic(err)
		}

		if MediaMapGet(a.Cover) == nil {
			err = fddb.SetAlbumCover(a.Id, "", "", "")
			if err != nil {
				panic(err)
			}
		}
	}
}

func GetAlbum(albumId types.AlbumId) (a *AlbumData, err error) {
	a, err = fddb.GetAlbum(albumId)
	return
}

func (a AlbumData) CanUserAccess(username types.Username) bool {
	return a.Owner == username || slices.Contains(a.SharedWith, username)
}

func (a *AlbumData) AddMedia(ms []types.Media) (addedCount int, err error) {
	mIds := util.Map(ms, func(m types.Media) types.ContentId { return m.Id() })
	a.Medias = util.AddToSet(a.Medias, mIds)
	addedCount, err = fddb.addMediaToAlbum(a.Id, mIds)
	return
}

func (a *AlbumData) Rename(newName string) (err error) {
	err = fddb.setAlbumName(a.Id, newName)
	return
}

func (a *AlbumData) RemoveMedia(ms []types.ContentId) (err error) {
	a.Medias = util.Filter(a.Medias, func(s types.ContentId) bool { return !slices.Contains(ms, s) })
	err = fddb.removeMediaFromAlbum(a.Id, ms)
	return
}

func (a *AlbumData) SetCover(mediaId types.ContentId) error {
	m := MediaMapGet(mediaId)
	if m == nil {
		return ErrNoMedia
	}
	colors, err := m.(*media).getProminentColors()
	if err != nil {
		return err
	}

	err = fddb.SetAlbumCover(a.Id, mediaId, colors[0], colors[1])
	if err != nil {
		return fmt.Errorf("failed to set album cover")
	}
	return nil
}

func (a *AlbumData) AddUsers(usernames []types.Username) (err error) {
	a.SharedWith = util.AddToSet(a.SharedWith, usernames)
	err = fddb.shareAlbum(a.Id, usernames)
	return
}

func (a *AlbumData) RemoveUsers(usernames []types.Username) (err error) {
	a.SharedWith = util.Filter(a.SharedWith, func(s types.Username) bool { return !slices.Contains(usernames, s) })
	err = fddb.unshareAlbum(a.Id, usernames)
	return
}
