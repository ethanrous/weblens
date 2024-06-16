package dataStore

import (
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func VerifyAlbumsMedia() {
	as := dbServer.getAllAlbums()
	if len(as) == 0 {
		return
	}

	for _, a := range as {
		deadMs := util.Filter(a.Medias, func(m types.ContentId) bool { return media.MediaMapGet(m) == nil })
		err := dbServer.removeMediaFromAlbum(a.Id, deadMs)
		if err != nil {
			panic(err)
		}

		if media.MediaMapGet(a.Cover) == nil {
			err = dbServer.SetAlbumCover(a.Id, "", "", "")
			if err != nil {
				panic(err)
			}
		}
	}
}

func GetAlbum(albumId types.AlbumId) (a *AlbumData, err error) {
	a, err = dbServer.GetAlbum(albumId)
	return
}

func DeleteAlbum(albumId types.AlbumId) (err error) {
	err = dbServer.DeleteAlbum(albumId)
	return
}

func (a *AlbumData) CanUserAccess(username types.Username) bool {
	return a.Owner == username || slices.Contains(a.SharedWith, username)
}

func (a *AlbumData) AddMedia(ms []types.Media) (addedCount int, err error) {
	mIds := util.Map(ms, func(m types.Media) types.ContentId { return m.ID() })
	a.Medias = util.AddToSet(a.Medias, mIds)
	addedCount, err = dbServer.addMediaToAlbum(a.Id, mIds)
	return
}

func (a *AlbumData) Rename(newName string) (err error) {
	err = dbServer.setAlbumName(a.Id, newName)
	return
}

func (a *AlbumData) RemoveMedia(ms []types.ContentId) (err error) {
	a.Medias = util.Filter(a.Medias, func(s types.ContentId) bool { return !slices.Contains(ms, s) })
	err = dbServer.removeMediaFromAlbum(a.Id, ms)
	return
}

func (a *AlbumData) SetCover(mediaId types.ContentId, ft types.FileTree) error {
	m := media.MediaMapGet(mediaId)
	if m == nil {
		return ErrNoMedia
	}
	colors, err := m.(*media.media).getProminentColors(ft)
	if err != nil {
		return err
	}

	err = dbServer.SetAlbumCover(a.Id, mediaId, colors[0], colors[1])
	if err != nil {
		return fmt.Errorf("failed to set album cover")
	}
	return nil
}

func (a *AlbumData) AddUsers(usernames []types.Username) (err error) {
	a.SharedWith = util.AddToSet(a.SharedWith, usernames)
	err = dbServer.shareAlbum(a.Id, usernames)
	return
}

func (a *AlbumData) RemoveUsers(usernames []types.Username) (err error) {
	a.SharedWith = util.Filter(a.SharedWith, func(s types.Username) bool { return !slices.Contains(usernames, s) })
	err = dbServer.unshareAlbum(a.Id, usernames)
	return
}
