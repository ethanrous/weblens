package dataStore

import (
	"fmt"
	"slices"

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
	m, err := MediaMapGet(mediaId)
	if err != nil {
		return err
	}
	colors, err := m.GetProminentColors()
	if err != nil {
		return err
	}

	err = fddb.SetAlbumCover(a.Id, mediaId, colors[0], colors[1])
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

func (a *AlbumData) CleanMissingMedia() (err error) {
	missing := util.Filter(a.Medias, func(s string) bool { _, err := MediaMapGet(s); return err == ErrNoMedia })
	if len(missing) != 0 {
		err = fddb.removeMediaFromAlbum(a.Id, missing)
		if err != nil {
			return
		}
	}
	a.Medias = util.Filter(a.Medias, func(s string) bool { return !slices.Contains(missing, s) })

	// If the cover is missing, reset to the first media, or none if no more media
	if !slices.Contains(a.Medias, a.Cover) {
		if len(a.Medias) != 0 {
			err = a.SetCover(a.Medias[0])
			if err != nil {
				return
			}
		} else {
			fddb.SetAlbumCover(a.Id, "", "", "")
		}
	}

	return
}
