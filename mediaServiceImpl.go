package weblens

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"golang.org/x/image/webp"

	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MediaServiceImpl struct {
	mediaMap    map[ContentId]*Media
	mapLock     sync.RWMutex
	typeService MediaTypeService
	exif        *exiftool.Exiftool
	mediaCache  *sturdyc.Client[[]byte]

	collection   *mongo.Collection
	fileService  *FileServiceImpl
	albumService *AlbumServiceImpl
}

func NewMediaService(
	fileService *FileServiceImpl, albumService *AlbumServiceImpl, mediaTypeServ MediaTypeService, col *mongo.Collection,
) *MediaServiceImpl {
	return &MediaServiceImpl{
		mediaMap:     make(map[ContentId]*Media),
		typeService: mediaTypeServ,
		exif:         newExif(1000*1000*100, 0, nil),
		mediaCache:  sturdyc.New[[]byte](1500, 10, time.Hour, 10),
		fileService:  fileService,
		albumService: albumService,
		collection:   col,
	}
}

func (ms *MediaServiceImpl) Init() werror.WErr {
	ret, err := ms.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return werror.Wrap(err)
	}

	var target []*Media
	err = ret.All(context.Background(), &target)
	if err != nil {
		return werror.Wrap(err)
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()

	for _, m := range target {
		ms.mediaMap[m.ID()] = m
	}

	return nil
}

func (ms *MediaServiceImpl) Size() int {
	return len(ms.mediaMap)
}

func (ms *MediaServiceImpl) Add(m *Media) werror.WErr {
	if m == nil {
		return werror.WErrMsg("attempt to set nil Media in map")
	}

	if m.ID() == "" {
		return werror.WErrMsg("Media id is empty")
	}

	if m.GetPageCount() == 0 {
		return werror.WErrMsg("Media page count is 0")
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()

	if ms.mediaMap[m.ID()] != nil {
		return werror.WErrMsg("attempt to re-add Media already in map")
	}

	if !m.IsImported() {
		m.SetImported(true)
		_, err := ms.collection.InsertOne(context.Background(), m)
		if err != nil {
			return werror.Wrap(err)
		}
	}

	ms.mediaMap[m.ID()] = m

	return nil
}

func (ms *MediaServiceImpl) TypeService() MediaTypeService {
	return ms.typeService
}

func (ms *MediaServiceImpl) Get(mId ContentId) *Media {
	if mId == "" {
		return nil
	}

	ms.mapLock.RLock()
	m := ms.mediaMap[mId]
	ms.mapLock.RUnlock()

	if reflect2.IsNil(m) {
		return nil
	}

	return m
}

func (ms *MediaServiceImpl) GetAll() []*Media {
	ms.mapLock.RLock()
	defer ms.mapLock.RUnlock()
	medias := internal.MapToValues(ms.mediaMap)
	return internal.SliceConvert[*Media](medias)
}

func (ms *MediaServiceImpl) Del(cId ContentId) error {
	m := ms.Get(cId)

	f, werr := m.GetCacheFile(Thumbnail, false, 0)
	if werr == nil {
		werr = ms.fileService.DeleteCacheFile(f)
		if werr != nil {
			return werr
		}
	}
	f = nil
	for page := range m.GetPageCount() + 1 {
		f, werr = m.GetCacheFile(Fullres, false, page)
		if werr == nil {
			werr = ms.fileService.DeleteCacheFile(f)
			if werr != nil {
				return werr
			}
		}
	}

	werr = ms.albumService.RemoveMediaFromAny(m.ID())
	if werr != nil {
		return werr
	}

	filter := bson.M{"contentId": m.ID()}
	_, err := ms.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return werror.Wrap(err)
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()
	delete(ms.mediaMap, m.ID())

	return nil
}

func (ms *MediaServiceImpl) FetchCacheImg(m *Media, q MediaQuality, pageNum int) ([]byte, error) {
	cacheKey := string(m.ID()) + string(q) + strconv.Itoa(pageNum)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "cacheKey", cacheKey)
	ctx = context.WithValue(ctx, "quality", q)
	ctx = context.WithValue(ctx, "pageNum", pageNum)
	ctx = context.WithValue(ctx, "media", m)

	cache, err := ms.mediaCache.GetFetch(ctx, cacheKey, types.SERV.StoreService.GetFetchMediaCacheImage)
	if err != nil {
		return nil, werror.Wrap(err)
	}
	return cache, nil
}

func (ms *MediaServiceImpl) StreamCacheVideo(m *Media, startByte, endByte int) ([]byte, error) {
	return nil, werror.NotImplemented("StreamCacheVideo")
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

func (ms *MediaServiceImpl) GetFilteredMedia(
	requester *User, sort string, sortDirection int, albumFilter []AlbumId,
	allowRaw bool, allowHidden bool,
) ([]*Media, error) {
	// old version
	// return dbServer.GetFilteredMedia(sort, requester.GetUsername(), -1, albumFilter, allowRaw)
	albums := internal.Map(
		albumFilter, func(aId AlbumId) *Album {
			return ms.albumService.Get(aId)
		},
	)

	var mediaMask []ContentId
	for _, a := range albums {
		mediaMask = append(
			mediaMask, internal.Map(
				a.GetMedias(), func(media *Media) ContentId {
					return media.ID()
				},
			)...,
		)
	}
	slices.Sort(mediaMask)

	ms.mapLock.RLock()
	allMs := internal.MapToValues(ms.mediaMap)
	ms.mapLock.RUnlock()
	allMs = internal.Filter(
		allMs, func(m *Media) bool {
			mt := m.GetMediaType()
			if mt == nil {
				return false
			}

			// Exclude Media if it is present in the filter
			_, e := slices.BinarySearch(mediaMask, m.ID())

			return !e &&
				m.GetOwner() == requester &&
				len(m.GetFiles()) != 0 &&
				(!mt.IsRaw() || allowRaw) &&
				(!m.IsHidden() || allowHidden) &&
				!mt.IsMime("application/pdf")
		},
	)

	slices.SortFunc(
		allMs, func(a, b *Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) * sortDirection },
	)

	return internal.SliceConvert[*Media](allMs), nil
}

func AdjustMediaDates(anchor *Media, newTime time.Time, extraMedias []*Media) error {
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

func (ms *MediaServiceImpl) RunExif(path string) ([]exiftool.FileMetadata, error) {
	if ms.exif == nil {
		return nil, werror.ErrNoExiftool
	}
	return ms.exif.ExtractMetadata(path), nil
}

func (ms *MediaServiceImpl) IsCached(m *Media) bool {
	if m.thumbCacheFile == nil {
		if m.ThumbnailCacheId == "" {
			return false
		}
		cache := ms.fileService.Get(m.ThumbnailCacheId)
		if cache == nil {
			return false
		}
		m.thumbCacheFile = cache
	}

	if len(m.fullresCacheFiles) == 0 {
		m.fullresCacheFiles = make([]*fileTree.WeblensFile, m.PageCount)
	}

	if m.fullresCacheFiles[0] == nil {
		if len(m.FullresCacheIds) != 0 {
			for i := range m.PageCount {
				cache := types.SERV.FileTree.Get(m.FullresCacheIds[i])
				if cache == nil {
					return false
				}
				m.fullresCacheFiles[i] = cache
			}
		}
	}

	return true
}

func (ms *MediaServiceImpl) GetProminentColors(media *Media) (prom []string, err error) {
	var i image.Image
	thumbBytes, err := ms.FetchCacheImg(media, Thumbnail, 0)
	if err != nil {
		return
	}

	i, err = webp.Decode(bytes.NewBuffer(thumbBytes))
	if err != nil {
		return
	}

	promColors, err := prominentcolor.Kmeans(i)
	prom = internal.Map(promColors, func(p prominentcolor.ColorItem) string { return p.AsString() })
	return
}

func (ms *MediaServiceImpl) NukeCache() error {
	ms.mapLock.Lock()
	cache := types.SERV.FileTree.Get("CACHE")
	for _, child := range cache.GetChildren() {
		err := types.SERV.FileTree.Del(child.ID())
		if err != nil {
			return err
		}
	}

	err := types.SERV.StoreService.DeleteAllMedia()
	if err != nil {
		return err
	}

	clear(ms.mediaMap)
	ms.mediaCache = sturdyc.New[[]byte](1500, 10, time.Hour, 10)

	ms.mapLock.Unlock()

	return nil
}

func fetchAndCacheMedia(ctx context.Context) (data []byte, err error) {
	defer internal.RecoverPanic("Failed to fetch media image into cache")

	m := ctx.Value("Media").(*Media)
	// util.Debug.Printf("Media cache miss [%s]", m.ID())

	q := ctx.Value("quality").(MediaQuality)
	pageNum := ctx.Value("pageNum").(int)

	f, err := m.GetCacheFile(q, true, pageNum)
	if err != nil {
		return
	}

	if f == nil {
		panic("This should never happen...")
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

func (ms *MediaServiceImpl) SetMediaLiked(mediaId ContentId, liked bool, username *Username) error {
	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()
	m, ok := ms.mediaMap[mediaId]
	if !ok {
		return werror.WErrMsg(fmt.Sprintf("Could not find media trying to like with id [%s]", mediaId))
	}

	err := ms.db.AddLikeToMedia(mediaId, username, liked)
	if err != nil {
		return err
	}

	if liked {
		m.LikedBy = internal.AddToSet(m.LikedBy, username)
	} else {
		m.LikedBy = internal.Filter(
			m.LikedBy, func(u *Username) bool {
				return u != username
			},
		)
	}

	return nil
}

func (ms *MediaServiceImpl) RemoveFileFromMedia(mediaId ContentId, fileId types.FileId) error {
	err := ms.db.RemoveFileFromMedia(mediaId, fileId)
	if err != nil {
		return err
	}

	return nil
}

func newExif(targetSize, currentSize int64, gexift *exiftool.Exiftool) *exiftool.Exiftool {
	if targetSize <= currentSize {
		return gexift
	}
	if gexift != nil {
		err := gexift.Close()
		wlog.ErrTrace(err)
		gexift = nil
	}
	buf := make([]byte, int(targetSize))
	et, err := exiftool.NewExiftool(
		exiftool.Api("largefilesupport"),
		exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(buf, int(targetSize)),
	)
	if err != nil {
		wlog.ErrTrace(err)
		return nil
	}
	gexift = et

	return gexift
}
