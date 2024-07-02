package media

import (
	"context"
	"fmt"
	"slices"
	"strconv"
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
	exif        *exiftool.Exiftool
	mediaCache  *sturdyc.Client[[]byte]
}

func NewRepo(mediaTypeServ types.MediaTypeService) types.MediaRepo {
	return &mediaRepo{
		mediaMap:    make(map[types.ContentId]types.Media),
		mapLock:     &sync.Mutex{},
		typeService: mediaTypeServ,
		exif:        NewExif(1000*1000*100, 0, nil),
		mediaCache:  sturdyc.New[[]byte](1500, 10, time.Hour, 10),
	}
}

func (mr *mediaRepo) Init(db types.DatabaseService) error {
	ms, err := db.GetAllMedia()
	if err != nil {
		return err
	}

	mr.mapLock.Lock()
	defer mr.mapLock.Unlock()

	for _, m := range ms {
		mr.mediaMap[m.ID()] = m
	}

	return nil
}

func (mr *mediaRepo) Size() int {
	return len(mr.mediaMap)
}

func (mr *mediaRepo) Add(m types.Media) error {
	if m == nil {
		return types.NewWeblensError("attempt to set nil Media in map")
	}

	if m.ID() == "" {
		return types.NewWeblensError("Media id is empty")
	}

	if m.GetPageCount() == 0 {
		return types.NewWeblensError("Media page count is 0")
	}

	mr.mapLock.Lock()
	defer mr.mapLock.Unlock()

	if mr.mediaMap[m.ID()] != nil {
		return types.NewWeblensError("attempt to re-add Media already in map")
	}

	if !m.IsImported() {
		m.SetImported(true)
		err := types.SERV.Database.CreateMedia(m)
		if err != nil {
			return err
		}
	}

	mr.mediaMap[m.ID()] = m

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

func (mr *mediaRepo) GetAll() []types.Media {
	mr.mapLock.Lock()
	defer mr.mapLock.Unlock()
	medias := util.MapToSlicePure(mr.mediaMap)
	return medias
}

func (mr *mediaRepo) Del(cId types.ContentId) error {
	m := mr.Get(cId)

	f, err := m.GetCacheFile(types.Thumbnail, false, 0)
	if err == nil {
		err = dataStore.PermanentlyDeleteFile(f, types.SERV.Caster)
		if err != nil {
			return err
		}
	}
	f = nil
	for page := range m.GetPageCount() + 1 {
		f, err = m.GetCacheFile(types.Fullres, false, page)
		if err == nil {
			err = dataStore.PermanentlyDeleteFile(f, types.SERV.Caster)
			if err != nil {
				return err
			}
		}
	}

	err = types.SERV.AlbumManager.RemoveMediaFromAny(m.ID())
	if err != nil {
		return err
	}

	err = types.SERV.Database.DeleteMedia(m.ID())
	if err != nil {
		return err
	}

	mr.mapLock.Lock()
	delete(mr.mediaMap, m.ID())
	mr.mapLock.Unlock()

	return nil
}

func (mr *mediaRepo) FetchCacheImg(m types.Media, q types.Quality, pageNum int) ([]byte, error) {
	cacheKey := string(m.ID()) + string(q) + strconv.Itoa(pageNum)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "cacheKey", cacheKey)
	ctx = context.WithValue(ctx, "quality", q)
	ctx = context.WithValue(ctx, "pageNum", pageNum)
	ctx = context.WithValue(ctx, "Media", m)

	cache, err := mr.mediaCache.GetFetch(ctx, cacheKey, fetchAndCacheMedia)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (mr *mediaRepo) StreamCacheVideo(m types.Media, startByte, endByte int) ([]byte, error) {
	return nil, types.ErrNotImplemented("StreamCacheVideo")
	// cacheKey := fmt.Sprintf("%s-STREAM %d-%d", m.ID(), startByte, endByte)

	// ctx := context.Background()
	// ctx = context.WithValue(ctx, "cacheKey", cacheKey)
	// ctx = context.WithValue(ctx, "startByte", startByte)
	// ctx = context.WithValue(ctx, "endByte", endByte)
	// ctx = context.WithValue(ctx, "Media", m)

	// video, err := fetchAndCacheVideo(m.(*Media), startByte, endByte)
	// if err != nil {
	// 	return nil, err
	// }
	// cache, err := mr.mediaCache.GetFetch(ctx, cacheKey, fetchAndCacheVideo)
	// if err != nil {
	// 	return nil, err
	// }
	// return cache, nil
}

func (mr *mediaRepo) GetFilteredMedia(
	requester types.User, sort string, sortDirection int, albumFilter []types.AlbumId,
	raw bool,
) ([]types.Media, error) {
	// old version
	// return dbServer.GetFilteredMedia(sort, requester.GetUsername(), -1, albumFilter, raw)
	albums := util.Map(
		albumFilter, func(aId types.AlbumId) types.Album {
			return types.SERV.AlbumManager.Get(aId)
		},
	)

	var mediaMask []types.ContentId
	for _, a := range albums {
		mediaMask = append(
			mediaMask, util.Map(
				a.GetMedias(), func(media types.Media) types.ContentId {
					return media.ID()
				},
			)...,
		)
	}
	slices.Sort(mediaMask)

	allMs := util.MapToSlicePure(mr.mediaMap)
	allMs = util.Filter(
		allMs, func(m types.Media) bool {
			mt := m.GetMediaType()
			if mt == nil {
				return false
			}

			// Exclude Media if it is present in the filter
			_, e := slices.BinarySearch(mediaMask, m.ID())

			return m.GetOwner() == requester && len(m.GetFiles()) != 0 && (!mt.IsRaw() || raw) && !mt.IsMime("application/pdf") && !e && !m.IsHidden()
		},
	)

	// Sort in timeline format, where most recent Media is at the beginning of the slice
	slices.SortFunc(allMs, func(a, b types.Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) * -1 })

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

func (mr *mediaRepo) RunExif(path string) ([]exiftool.FileMetadata, error) {
	if mr.exif == nil {
		return nil, types.ErrNoExiftool
	}
	return mr.exif.ExtractMetadata(path), nil
}

func fetchAndCacheMedia(ctx context.Context) (data []byte, err error) {
	defer util.RecoverPanic("Failed to fetch media image into cache")

	m := ctx.Value("Media").(*Media)
	// util.Debug.Printf("Media cache miss [%s]", m.ID())

	q := ctx.Value("quality").(types.Quality)
	pageNum := ctx.Value("pageNum").(int)

	f, err := m.GetCacheFile(q, true, pageNum)
	if err != nil {
		return
	}

	if f == nil {
		return nil, types.ErrNoFile
	}

	data, err = f.ReadAll()
	if err != nil {
		return
	}
	if len(data) == 0 {
		err = fmt.Errorf("displayable bytes empty")
		return
	}
	return
}
