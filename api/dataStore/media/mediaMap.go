package media

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type mediaRepo struct {
	mediaMap    map[types.ContentId]types.Media
	mapLock     *sync.Mutex
	typeService types.MediaTypeService
	imageCache  *sturdyc.Client[[]byte]
	exif        *exiftool.Exiftool
	db          types.DatabaseService
}

func NewRepo(service types.MediaTypeService, db types.DatabaseService) types.MediaRepo {
	return &mediaRepo{
		mediaMap:    make(map[types.ContentId]types.Media),
		mapLock:     &sync.Mutex{},
		typeService: service,
		imageCache:  sturdyc.New[[]byte](500, 10, time.Hour, 10),
		exif:        NewExif(10*100*100, 0, nil),
		db:          db,
	}
}

func (mr *mediaRepo) Init() error {
	db.
	_, err := dataStore.dbServer.getAllMedia()
	if err != nil {
		panic(err)
	}

	// for _, m := range ms {
	// 	mediaMapAdd(m)
	// }

	return nil
}

func (mr *mediaRepo) Size() int {
	return len(mr.mediaMap)
}

func (mr *mediaRepo) Add(m types.Media) error {
	if m == nil {
		return types.NewWeblensError("attempt to set nil Media in map")
	}
	if !m.IsImported() {
		return types.NewWeblensError("tried adding non-imported Media to map")
	}

	if m.GetPageCount() == 0 {
		return types.NewWeblensError("Media page count is 0")
		// m. = 1
		// err := dataStore.dbServer.UpdateMedia(m)
		// if err != nil {
		// 	util.ErrTrace(err)
		// }
	}

	mr.mapLock.Lock()
	defer mr.mapLock.Unlock()

	if mr.mediaMap[m.ID()] != nil {
		return types.NewWeblensError("attempt to re-add Media already in map")
	}

	// if m.fullresCacheFiles == nil || len(m.fullresCacheFiles) < m.PageCount {
	// 	m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
	// }
	// if m.FullresCacheIds == nil || len(m.FullresCacheIds) < m.PageCount {
	// 	m.FullresCacheIds = make([]types.FileId, m.PageCount)
	// }
	// if m.mediaType == nil {
	// 	m.mediaType = mr.typeService.ParseMime(m)
	// }

	mr.mediaMap[m.ID()] = m
	mr.mapLock.Unlock()

	// orphaned := true
	// for _, fId := range m.FileIds {
	// 	f := ft.Get(fId)
	// 	if f == nil {
	// 		m.RemoveFile(fId)
	// 		continue
	// 	}
	// 	orphaned = false
	// 	// f.SetMedia(m)
	// }
	// if orphaned && len(m.FileIds) != 0 {
	// 	removeMedia(m, ft)
	// }
	return nil
}

func (mr *mediaRepo) TypeService() types.MediaTypeService {
	return mr.typeService
}

func (mr *mediaRepo) Get(mId types.ContentId) types.Media {
	if mId == "" {
		return nil
	}

	mr.mapLock.Lock()
	m := mr.mediaMap[mId]
	mr.mapLock.Unlock()

	return m
}

func (mr *mediaRepo) Del(m types.Media, ft types.FileTree) {

	realM := m.(*Media)
	f, err := realM.getCacheFile(dataStore.Thumbnail, false, 0, ft)
	if err == nil {
		err = dataStore.PermanentlyDeleteFile(f, nil)
		if err != nil {
			util.ErrTrace(err)
		}
	}
	f = nil
	for page := range realM.PageCount + 1 {
		f, err = realM.getCacheFile(dataStore.Fullres, false, page, ft)
		if err == nil {
			err = dataStore.PermanentlyDeleteFile(f, nil)
			if err != nil {
				util.ErrTrace(err)
			}
		}
	}

	err = dataStore.dbServer.removeMediaFromAnyAlbum(m.ID())
	if err != nil {
		util.ErrTrace(err)
		return
	}

	err = dataStore.dbServer.deleteMedia(m.ID())
	if err != nil {
		util.ErrTrace(err)
		return
	}

	mr.mapLock.Lock()
	delete(mr.mediaMap, m.ID())
	mr.mapLock.Unlock()
}

func (mr *mediaRepo) FetchCacheImg(m types.Media) error {
	if mr.Get(m.ID()) == nil {
		return ErrNoMedia
	}

}

func GetRealFile(m types.Media, ft types.FileTree) (types.WeblensFile, error) {
	realM := m.(*Media)

	if len(realM.FileIds) == 0 {
		return nil, dataStore.ErrNoFile
	}

	for _, fId := range realM.FileIds {
		f := ft.Get(fId)
		if f != nil {
			return f, nil
		}
	}

	// None of the files that this Media uses are present any longer, delete Media
	removeMedia(realM, ft)
	return nil, dataStore.ErrNoFile
}

// func GetRandomMedia(limit int) []types.Media {
// 	count := 0
// 	medias := []types.Media{}
// 	for _, m := range mediaMap {
// 		if count == limit {
// 			break
// 		}
// 		if m.GetPageCount() != 1 {
// 			continue
// 		}
// 		medias = append(medias, m)
// 		count++
// 	}
//
// 	return medias
// }

func sortMediaByOwner(a, b types.Media) int {
	return strings.Compare(string(a.GetOwner().GetUsername()), string(b.GetOwner().GetUsername()))
}

func findOwner(m types.Media, o types.User) int {
	return strings.Compare(string(m.GetOwner().GetUsername()), string(o.GetUsername()))
}

func GetFilteredMedia(requester types.User, sort string, sortDirection int, albumFilter []types.AlbumId, raw bool) ([]types.Media, error) {
	// old version
	// return dbServer.GetFilteredMedia(sort, requester.GetUsername(), -1, albumFilter, raw)
	albums := util.Map(albumFilter, func(a types.AlbumId) *dataStore.AlbumData {
		album, err := dataStore.dbServer.GetAlbum(a)
		util.ShowErr(err)
		return album
	})

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

		// Exclude Media if it is present in the filter
		_, e := slices.BinarySearch(mediaMask, m.ID())

		return m.GetOwner() == requester && len(m.GetFiles()) != 0 && (!mt.IsRaw() || raw) && !mt.IsMime("application/pdf") && !e && !m.IsHidden()
	})

	// Sort in timeline format, where most recent Media is at the beginning of the slice
	slices.SortFunc(allMs, func(a, b types.Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) })

	return allMs, nil
}

func AdjustMediaDates(anchor types.Media, newTime time.Time, extraMedias []types.Media) error {
	offset := newTime.Sub(anchor.GetCreateDate())

	err := anchor.SetCreateDate(anchor.GetCreateDate().Add(offset))
	if err != nil {
		return err
	}
	for _, m := range extraMedias {
		err = m.SetCreateDate(m.GetCreateDate().Add(offset))
		if err != nil {
			return err
		}
	}

	return nil
}

// Error

var ErrNoMedia = errors.New("requested Media was not found in Media repo")
var ErrNoFile = errors.New("Media does not have an associated file that it was expected to")
