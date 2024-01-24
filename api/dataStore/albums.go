package dataStore

import (
	"fmt"
	"slices"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethrousseau/weblens/api/util"
)

func GetAlbum(albumId string) (a *AlbumData, err error) {
	a, err = fddb.GetAlbum(albumId)
	return
}

func (a AlbumData) CanUserAccess(username string) bool {
	return a.Owner == username || slices.Contains(a.SharedWith, username)
}

func (a *AlbumData) AddMedia(ms []string) (addedCount int, err error) {
	a.Medias = util.AddToSet(a.Medias, ms)
	addedCount, err = fddb.addMediaToAlbum(a.Id, ms)
	return
}

func (a *AlbumData) Rename(newName string) (err error) {
	err = fddb.setAlbumName(a.Id, newName)
	return
}

func (a *AlbumData) RemoveMedia(ms []string) (err error) {
	a.Medias = util.Filter(a.Medias, func(s string) bool { return !slices.Contains(ms, s) })
	err = fddb.removeMediaFromAlbum(a.Id, ms)
	return
}

func (a *AlbumData) SetCover(mediaId string) error {
	m := fddb.GetMedia(mediaId)
	i := m.GetImage()
	prom, err := prominentcolor.Kmeans(i)
	if err != nil {
		return err
	}

	err = fddb.SetAlbumCover(a.Id, mediaId, prom[0].AsString(), prom[1].AsString())
	if err != nil {
		return fmt.Errorf("failed to set album cover")
	}
	return nil
}

func (a *AlbumData) AddUsers(usernames []string) (err error) {
	a.SharedWith = util.AddToSet(a.SharedWith, usernames)
	err = fddb.shareAlbum(a.Id, usernames)
	return
}

func (a *AlbumData) RemoveUsers(usernames []string) (err error) {
	a.SharedWith = util.Filter(a.SharedWith, func(s string) bool { return !slices.Contains(usernames, s) })
	err = fddb.unshareAlbum(a.Id, usernames)
	return
}
