package dataStore

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var mediaMap map[types.ContentId]types.Media = map[types.ContentId]types.Media{}
var mediaMapLock *sync.Mutex = &sync.Mutex{}

func MediaInit() error {
	_, err := fddb.getAllMedia()
	if err != nil {
		return err
	}

	// for _, m := range ms {
	// 	mediaMapAdd(m)
	// }

	return nil
}

func GetMediaMapSize() int {
	return len(mediaMap)
}

func mediaMapAdd(m *media) {
	if m == nil {
		util.ErrTrace(fmt.Errorf("attempt to set nil media in map"))
		return
	}
	if !m.IsImported() {
		util.ErrTrace(fmt.Errorf("tried adding non-imported media to map"))
		return
	}

	mediaMapLock.Lock()

	if mediaMap[m.ContentId] != nil {
		mediaMapLock.Unlock()
		util.Error.Println(fmt.Errorf("attempt to re-add media already in map"))
		return
	}

	if m.PageCount == 0 {
		m.PageCount = 1
		err := fddb.UpdateMedia(m)
		if err != nil {
			util.ErrTrace(err)
		}
	}

	if m.fullresCacheFiles == nil {
		m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
	}
	if m.FullresCacheIds == nil {
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
	}
	if m.mediaType == nil {
		m.mediaType = ParseMimeType(m.MimeType)
	}

	mediaMap[m.Id()] = m

	mediaMapLock.Unlock()

	orphaned := true
	for _, fId := range m.FileIds {
		f := FsTreeGet(fId)
		if f == nil {
			m.RemoveFile(fId)
			continue
		}
		orphaned = false
		// f.SetMedia(m)
	}
	if orphaned && len(m.FileIds) != 0 {
		removeMedia(m)
	}
}

func MediaMapGet(mId types.ContentId) types.Media {
	if mId == "" {
		return nil
	}

	mediaMapLock.Lock()
	m := mediaMap[mId]
	mediaMapLock.Unlock()

	return m
}

func removeMedia(m types.Media) {

	util.Error.Println("actually no")
	return

	realM := m.(*media)
	f, err := realM.getCacheFile(Thumbnail, false, 0)
	if err == nil {
		PermanentlyDeleteFile(f, voidCaster)
	}
	f = nil
	for page := range realM.PageCount + 1 {
		f, err = realM.getCacheFile(Fullres, false, page)
		if err == nil {
			PermanentlyDeleteFile(f, voidCaster)
		}
	}

	err = fddb.deleteMedia(m.Id())
	if err != nil {
		util.ErrTrace(err)
		return
	}

	mediaMapLock.Lock()
	delete(mediaMap, m.Id())
	mediaMapLock.Unlock()
}

func HideMedia(m types.Media) error {
	realM := m.(*media)
	realM.Hidden = true

	return fddb.setMediaHidden(m, true)
}

func GetRealFile(m types.Media) (types.WeblensFile, error) {
	realM := m.(*media)
	for _, fId := range realM.FileIds {
		f := FsTreeGet(fId)
		if f != nil {
			return f, nil
		}
	}

	// None of the files that this media uses are present any longer, delete media
	removeMedia(realM)
	return nil, ErrNoFile
}

func GetRandomMedia(limit int) []types.Media {
	count := 0
	medias := []types.Media{}
	for _, m := range mediaMap {
		if count == limit {
			break
		}
		if m.GetPageCount() != 1 {
			continue
		}
		medias = append(medias, m)
		count++
	}

	return medias
}

func sortMediaByOwner(a, b types.Media) int {
	return strings.Compare(string(a.GetOwner().GetUsername()), string(b.GetOwner().GetUsername()))
}

func findOwner(m types.Media, o types.User) int {
	return strings.Compare(string(m.GetOwner().GetUsername()), string(o.GetUsername()))
}

func GetFilteredMedia(requester types.User, sort string, sortDirection int, albumFilter []types.AlbumId, raw bool) ([]types.Media, error) {
	// old version
	// return fddb.GetFilteredMedia(sort, requester.GetUsername(), -1, albumFilter, raw)
	albums := util.Map(albumFilter, func(a types.AlbumId) *AlbumData { album, err := fddb.GetAlbum(a); util.ShowErr(err); return album })

	mediaMask := []types.ContentId{}
	for _, a := range albums {
		mediaMask = append(mediaMask, a.Medias...)
	}
	slices.Sort(mediaMask)

	allMs := util.MapToSlicePure(mediaMap)
	allMs = util.Filter(allMs, func(m types.Media) bool {
		mt := m.GetMediaType()
		if mt == nil {
			return false
		}

		// Exclude media if it is present in the filter
		_, e := slices.BinarySearch(mediaMask, m.Id())

		return len(m.GetFiles()) != 0 && (!mt.IsRaw() || raw) && mt.FriendlyName() != "PDF" && !e
	})

	allMs = util.Filter(allMs, func(m types.Media) bool {
		return m.GetOwner() == requester && !m.IsHidden()
	})

	// Sort in timeline format, where most recent media is at the beginning of the slice
	slices.SortFunc(allMs, func(a, b types.Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) })

	return allMs, nil
}
