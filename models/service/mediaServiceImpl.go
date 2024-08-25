package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/pkg/errors"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/webp"

	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.MediaService = (*MediaServiceImpl)(nil)

type MediaServiceImpl struct {
	mediaMap    map[models.ContentId]*models.Media
	streamerMap map[models.ContentId]*models.VideoStreamer
	mapLock     sync.RWMutex
	exif        *exiftool.Exiftool
	mediaCache  *sturdyc.Client[[]byte]

	typeService  models.MediaTypeService
	fileService  *FileServiceImpl
	albumService *AlbumServiceImpl

	collection *mongo.Collection
}

func NewMediaService(
	fileService *FileServiceImpl, albumService *AlbumServiceImpl, mediaTypeServ models.MediaTypeService,
	col *mongo.Collection,
) *MediaServiceImpl {
	return &MediaServiceImpl{
		mediaMap:     make(map[models.ContentId]*models.Media),
		typeService:  mediaTypeServ,
		exif:         newExif(1000*1000*100, 0, nil),
		mediaCache:   sturdyc.New[[]byte](1500, 10, time.Hour, 10),
		fileService:  fileService,
		albumService: albumService,
		collection:   col,
	}
}

func (ms *MediaServiceImpl) Init() error {
	ret, err := ms.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return werror.WithStack(err)
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()

	cursorContext := context.Background()
	for ret.Next(cursorContext) {
		m := &models.Media{}
		err = ret.Decode(m)
		if err != nil {
			return werror.WithStack(err)
		}
		ms.mediaMap[m.ID()] = m
	}

	return nil
}

func (ms *MediaServiceImpl) Size() int {
	return len(ms.mediaMap)
}

func (ms *MediaServiceImpl) Add(m *models.Media) error {
	if m == nil {
		return werror.WithStack(werror.New("attempt to set nil Media in map"))
	}

	if m.ID() == "" {
		return werror.WithStack(werror.New("Media id is empty"))
	}

	if m.GetPageCount() == 0 {
		return werror.WithStack(werror.New("Media page count is 0"))
	}

	if m.MediaWidth == 0 || m.MediaHeight == 0 {
		return werror.WithStack(werror.New("Media has unset dimentions"))
	}

	if len(m.FileIds) == 0 {
		return werror.WithStack(werror.New("Media has no files"))
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()

	if ms.mediaMap[m.ID()] != nil {
		return werror.New("attempt to re-add Media already in map")
	}

	if !m.IsImported() {
		m.SetImported(true)
		_, err := ms.collection.InsertOne(context.Background(), m)
		if err != nil {
			return werror.WithStack(err)
		}
	}

	ms.mediaMap[m.ID()] = m

	return nil
}

func (ms *MediaServiceImpl) TypeService() models.MediaTypeService {
	return ms.typeService
}

func (ms *MediaServiceImpl) Get(mId models.ContentId) *models.Media {
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

func (ms *MediaServiceImpl) GetAll() []*models.Media {
	ms.mapLock.RLock()
	defer ms.mapLock.RUnlock()
	medias := internal.MapToValues(ms.mediaMap)
	return internal.SliceConvert[*models.Media](medias)
}

func (ms *MediaServiceImpl) Del(cId models.ContentId) error {
	m := ms.Get(cId)

	f, werr := ms.getCacheFile(m, models.LowRes, 0)
	if werr == nil {
		werr = ms.fileService.DeleteCacheFile(f)
		if werr != nil {
			return werr
		}
	}
	f = nil
	for page := range m.GetPageCount() + 1 {
		f, werr = ms.getCacheFile(m, models.HighRes, page)
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
		return werror.WithStack(err)
	}

	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()
	delete(ms.mediaMap, m.ID())

	return nil
}

func (ms *MediaServiceImpl) HideMedia(m *models.Media, hidden bool) error {
	return werror.NotImplemented("HideMedia")
}

func (ms *MediaServiceImpl) FetchCacheImg(m *models.Media, q models.MediaQuality, pageNum int) ([]byte, error) {
	cacheKey := string(m.ID()) + string(q) + strconv.Itoa(pageNum)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "cacheKey", cacheKey)
	ctx = context.WithValue(ctx, "quality", q)
	ctx = context.WithValue(ctx, "pageNum", pageNum)
	ctx = context.WithValue(ctx, "media", m)

	cache, err := ms.mediaCache.GetFetch(ctx, cacheKey, ms.getFetchMediaCacheImage)
	if err != nil {
		return nil, werror.WithStack(err)
	}
	return cache, nil
}

func (ms *MediaServiceImpl) StreamCacheVideo(m *models.Media, startByte, endByte int) ([]byte, error) {
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
	requester *models.User, sort string, sortDirection int, albumFilter []models.AlbumId,
	allowRaw bool, allowHidden bool,
) ([]*models.Media, error) {
	albums := internal.Map(
		albumFilter, func(aId models.AlbumId) *models.Album {
			return ms.albumService.Get(aId)
		},
	)

	var mediaMask []models.ContentId
	for _, a := range albums {
		mediaMask = append(mediaMask, a.Medias...)
	}
	slices.Sort(mediaMask)

	ms.mapLock.RLock()
	allMs := internal.MapToValues(ms.mediaMap)
	ms.mapLock.RUnlock()
	allMs = internal.Filter(
		allMs, func(m *models.Media) bool {
			mt := ms.GetMediaType(m)
			if mt.Mime == "" || (mt.IsRaw() && !allowRaw) || (m.IsHidden() && !allowHidden) || m.GetOwner() != requester.GetUsername() || len(m.GetFiles()) == 0 || mt.IsMime("application/pdf") {
				return false
			}

			// Exclude Media if it is present in the filter
			_, e := slices.BinarySearch(mediaMask, m.ID())
			return !e
		},
	)

	slices.SortFunc(
		allMs, func(a, b *models.Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) * sortDirection },
	)

	return allMs, nil
}

func AdjustMediaDates(anchor *models.Media, newTime time.Time, extraMedias []*models.Media) error {
	offset := newTime.Sub(anchor.GetCreateDate())

	anchor.SetCreateDate(anchor.GetCreateDate().Add(offset))

	for _, m := range extraMedias {
		m.SetCreateDate(m.GetCreateDate().Add(offset))
	}

	// TODO - update media date in DB

	return nil
}

func (ms *MediaServiceImpl) IsCached(m *models.Media) bool {
	if m.ThumbnailCacheId == "" {
		return false
	}
	cache, err := ms.fileService.GetThumbFileId(m.ThumbnailCacheId)
	if err != nil {
		return false
	}

	if len(m.HighResCacheIds) != 0 {
		for i := range m.PageCount {
			cache, err = ms.fileService.GetThumbFileId(m.HighResCacheIds[i])
			if cache == nil {
				return false
			}
		}
	}

	return true
}

func (ms *MediaServiceImpl) IsFileDisplayable(f *fileTree.WeblensFile) bool {
	ext := filepath.Ext(f.Filename())
	return ms.typeService.ParseExtension(ext).Displayable
}

func (ms *MediaServiceImpl) GetProminentColors(media *models.Media) (prom []string, err error) {
	var i image.Image
	thumbBytes, err := ms.FetchCacheImg(media, models.LowRes, 0)
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
	// ms.mapLock.Lock()
	// ms.fileService.clearCacheDir()
	// cache := types.SERV.FileTree.Get("CACHE")
	// for _, child := range cache.GetChildren() {
	// 	err := types.SERV.FileTree.Del(child.ID())
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// err := types.SERV.StoreService.DeleteAllMedia()
	// if err != nil {
	// 	return err
	// }
	//
	// clear(ms.mediaMap)
	// ms.mediaCache = sturdyc.New[[]byte](1500, 10, time.Hour, 10)
	//
	// ms.mapLock.Unlock()
	return werror.NotImplemented("NukeCache")

	return nil
}

func (ms *MediaServiceImpl) StreamVideo(
	m *models.Media, u *models.User, share *models.FileShare,
) (*models.VideoStreamer, error) {
	var streamer *models.VideoStreamer
	var ok bool
	if streamer, ok = ms.streamerMap[m.ID()]; !ok {
		streamPath := fmt.Sprintf("%s/%s-stream/", internal.GetThumbsDir(), m.ID())
		streamer = models.NewVideoStreamer(m, streamPath)
		ms.streamerMap[m.ID()] = streamer
	}

	f, err := ms.fileService.GetFileSafe(m.FileIds[0], u, share)
	if err != nil {
		return nil, err
	}

	streamer.Encode(f)

	return streamer, nil
}

func (ms *MediaServiceImpl) SetMediaLiked(mediaId models.ContentId, liked bool, username models.Username) error {
	ms.mapLock.Lock()
	defer ms.mapLock.Unlock()
	m, ok := ms.mediaMap[mediaId]
	if !ok {
		return errors.Errorf("Could not find media with id [%s] while trying to update liked array", mediaId)
	}

	filter := bson.M{"contentId": mediaId}
	var update bson.M
	if liked {
		update = bson.M{"$addToSet": bson.M{"likedBy": username}}
	} else {
		update = bson.M{"$pull": bson.M{"likedBy": username}}
	}
	_, err := ms.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if liked {
		m.LikedBy = internal.AddToSet(m.LikedBy, username)
	} else {
		m.LikedBy = internal.Filter(
			m.LikedBy, func(u models.Username) bool {
				return u != username
			},
		)
	}

	return nil
}

func (ms *MediaServiceImpl) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {
	filter := bson.M{"contentId": media.ID()}
	update := bson.M{"$pull": bson.M{"fileIds": fileId}}
	_, err := ms.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	media.FileIds = internal.Filter(
		media.FileIds, func(fId fileTree.FileId) bool {
			return fId != fileId
		},
	)

	return nil
}

func (ms *MediaServiceImpl) LoadMediaFromFile(m *models.Media, file *fileTree.WeblensFile) error {
	fileMetas := ms.exif.ExtractMetadata(file.GetAbsPath())
	for _, fileMeta := range fileMetas {
		if fileMeta.Err != nil {
			return fileMeta.Err
		}
	}

	rawExif := fileMetas[0].Fields

	var err error
	if m.CreateDate.Unix() <= 0 {
		r, ok := rawExif["SubSecCreateDate"]
		if !ok {
			r, ok = rawExif["MediaCreateDate"]
		}
		if ok {
			m.CreateDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
			if err != nil {
				m.CreateDate, err = time.Parse("2006:01:02 15:04:05.00-07:00", r.(string))
			}
			if err != nil {
				m.CreateDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
			}
			if err != nil {
				m.CreateDate, err = time.Parse("2006:01:02 15:04:05-07:00", r.(string))
			}
			if err != nil {
				m.CreateDate = file.ModTime()
			}
		} else {
			m.CreateDate = file.ModTime()
		}
	}

	if m.MimeType == "" {
		mimeType, ok := rawExif["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MimeType = mimeType

		if ms.typeService.ParseMime(m.MimeType).IsVideo() {
			probeJson, err := ffmpeg.Probe(file.GetAbsPath())
			if err != nil {
				return err
			}
			probeResult := map[string]any{}
			err = json.Unmarshal([]byte(probeJson), &probeResult)
			if err != nil {
				return err
			}

			formatChunk, ok := probeResult["format"].(map[string]any)
			if !ok {
				return errors.New("invalid movie format")
			}
			duration, err := strconv.ParseFloat(formatChunk["duration"].(string), 10)
			if err != nil {
				return err
			}
			m.VideoLength = int(duration * 1000)
		}
	}

	if ms.typeService.ParseMime(m.MimeType).IsMultiPage() {
		m.PageCount = int(rawExif["PageCount"].(float64))
	} else {
		m.PageCount = 1
	}

	if len(m.HighResCacheIds) != m.PageCount {
		m.HighResCacheIds = make([]fileTree.FileId, m.PageCount)
	}

	if m.Rotate == "" {
		rotate := rawExif["Orientation"]
		if rotate != nil {
			m.Rotate = rotate.(string)
		}
	}

	var bs []byte

	mType := ms.GetMediaType(m)

	if mType.IsRaw() {
		raw64 := rawExif[mType.GetThumbExifKey()].(string)
		raw64 = raw64[strings.Index(raw64, ":")+1:]

		imgBytes, err := base64.StdEncoding.DecodeString(raw64)
		if err != nil {
			return err
		}
		bs = imgBytes
	} else if mType.IsVideo() {
		out := bytes.NewBuffer(nil)
		errOut := bytes.NewBuffer(nil)

		const frameNum = 10

		err = ffmpeg.Input(file.GetAbsPath()).Filter(
			"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
		).Output(
			"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
		).WithOutput(out).WithErrorOutput(errOut).Run()
		if err != nil {
			log.Error.Println(errOut.String())
			return werror.WithStack(err)
		}
		bs = out.Bytes()

	} else {
		bs, err = file.ReadAll()
		if err != nil {
			return err
		}
	}

	_, err = ms.generateCacheFiles(m, bs)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MediaServiceImpl) GetMediaType(m *models.Media) models.MediaType {
	return ms.typeService.ParseMime(m.MimeType)
}

func (ms *MediaServiceImpl) GetMediaTypes() models.MediaTypeService {
	return ms.typeService
}

func (ms *MediaServiceImpl) RecursiveGetMedia(folders ...*fileTree.WeblensFile) []models.ContentId {
	var medias []models.ContentId

	for _, f := range folders {
		if f == nil {
			log.Warning.Println("Skipping recursive media lookup for non-existent folder")
			continue
		}
		if !f.IsDir() {
			if ms.IsFileDisplayable(f) {
				m := ms.Get(models.ContentId(f.GetContentId()))
				if m != nil {
					medias = append(medias, m.ID())
				}
			}
			continue
		}
		err := f.RecursiveMap(
			func(f *fileTree.WeblensFile) error {
				if !f.IsDir() && ms.IsFileDisplayable(f) {
					m := ms.Get(models.ContentId(f.GetContentId()))
					if m != nil {
						medias = append(medias, m.ID())
					}
				}
				return nil
			},
		)
		if err != nil {
			log.ShowErr(err)
		}
	}

	return medias
}

func (ms *MediaServiceImpl) getFetchMediaCacheImage(ctx context.Context) ([]byte, error) {
	defer internal.RecoverPanic("Failed to fetch media image into cache")

	m := ctx.Value("media").(*models.Media)

	q := ctx.Value("quality").(models.MediaQuality)
	pageNum := ctx.Value("pageNum").(int)

	f, err := ms.getCacheFile(m, q, pageNum)
	if err != nil {
		return nil, err
		// caches, err := ms.generateCacheFiles(m)
		// if err != nil {
		// 	return nil, err
		// }
		//
		// if len(caches) <= pageNum+1 {
		// 	return nil, werror.ErrNoCache
		// }
		//
		// if q == LowRes {
		// 	f = caches[0]
		// } else {
		// 	f = caches[pageNum+1]
		// }
	}

	if f == nil {
		panic("This should never happen... file is nil in GetFetchMediaCacheImage")
	}

	data, err := f.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		err = errors.New("displayable bytes empty")
		return nil, err
	}
	return data, nil
}

func (ms *MediaServiceImpl) getCacheFile(
	m *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFile, error) {
	var pageNumStr string
	if pageNum != 0 {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}
	filename := fmt.Sprintf("%s-%s%s.cache", m.ID(), quality, pageNumStr)

	cacheFile, err := ms.fileService.GetThumbFileName(filename)
	if err != nil {
		return nil, werror.WithStack(werror.ErrNoCache)
	}
	return cacheFile, nil
}

func (ms *MediaServiceImpl) generateCacheFiles(m *models.Media, bs []byte) ([]*fileTree.WeblensFile, error) {
	// thumbFile, err := ms.getCacheFile(m, models.LowRes, 0)
	// if err == nil {
	// 	return []*fileTree.WeblensFile{thumbFile}, nil
	// }
	var err error

	img := bimg.NewImage(bs)

	if ms.GetMediaType(m).IsRaw() {
		switch m.Rotate {
		case "Rotate 270 CW":
			_, err = img.Rotate(270)
		case "Rotate 90 CW":
			_, err = img.Rotate(90)
		case "Horizontal (normal)":
		case "":
			log.Debug.Println("empty orientation")
		default:
			err = werror.Errorf("Unknown rotate name [%s]", m.Rotate)
		}
		if err != nil {
			return nil, werror.WithStack(err)
		}
	}

	_, err = img.Convert(bimg.WEBP)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	imgSize, err := img.Size()
	if err != nil {
		return nil, werror.WithStack(err)
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	thumbW := int((models.ThumbnailHeight / float32(m.MediaHeight)) * float32(m.MediaWidth))

	var cacheFiles []*fileTree.WeblensFile

	mType := ms.GetMediaType(m)
	if !mType.IsMultiPage() {

		// Copy image buffer for the thumbnail
		thumbImg := bimg.NewImage(img.Image())

		thumbBytes, err := thumbImg.Resize(thumbW, int(models.ThumbnailHeight))
		if err != nil {
			return nil, err
		}
		thumbSize, err := thumbImg.Size()
		if err != nil {
			return nil, werror.WithStack(err)
		} else {
			thumbRatio := float64(thumbSize.Width) / float64(thumbSize.Height)
			mediaRatio := float64(m.MediaWidth) / float64(m.MediaHeight)
			if (thumbRatio < 1 && mediaRatio > 1) || (thumbRatio > 1 && mediaRatio < 1) {
				log.Error.Println("Mismatched media sizes")
			}
		}

		var thumbFile *fileTree.WeblensFile
		thumbFile, err = ms.fileService.NewCacheFile(string(m.ID()), models.LowRes, 0)
		if err != nil {
			if !errors.Is(err, werror.ErrFileAlreadyExists) {
				return nil, err
			}
		} else {
			err = thumbFile.Write(thumbBytes)
			if err != nil {
				return nil, err
			}

			cacheFiles = append(cacheFiles, thumbFile)
		}

		var fullresFile *fileTree.WeblensFile
		fullresFile, err = ms.fileService.NewCacheFile(string(m.ID()), models.HighRes, 0)
		if err != nil {
			if !errors.Is(err, werror.ErrFileAlreadyExists) {
				return nil, err
			}
		} else {
			err = fullresFile.Write(img.Image())
			if err != nil {
				return nil, err
			}

			cacheFiles = append(cacheFiles, fullresFile)
		}

	}

	return cacheFiles, nil
}

func newExif(targetSize, currentSize int64, gexift *exiftool.Exiftool) *exiftool.Exiftool {
	if targetSize <= currentSize {
		return gexift
	}
	if gexift != nil {
		err := gexift.Close()
		log.ErrTrace(err)
		gexift = nil
	}
	buf := make([]byte, int(targetSize))
	et, err := exiftool.NewExiftool(
		exiftool.Api("largefilesupport"),
		exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(buf, int(targetSize)),
	)
	if err != nil {
		log.ErrTrace(err)
		return nil
	}
	gexift = et

	return gexift
}
