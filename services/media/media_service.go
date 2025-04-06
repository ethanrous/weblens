package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/davidbyttow/govips/v2/vips"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"image"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"slices"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"
//
// 	"github.com/EdlinOrg/prominentcolor"
// 	"github.com/ethanrous/weblens/fileTree"
// 	"github.com/ethanrous/weblens/internal"
// 	"github.com/pkg/errors"
// 	media_model "github.com/ethanrous/weblens/models/media"
// 	wl_slices "github.com/ethanrous/weblens/modules/slices"
// 	"github.com/rs/zerolog"
// 	ffmpeg "github.com/u2takey/ffmpeg-go"
// 	"github.com/viccon/sturdyc"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// 	"golang.org/x/image/webp"
//
// 	ollama "github.com/ollama/ollama/api"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
//
// 	"github.com/davidbyttow/govips/v2/vips"
// )

func GetConverted(m *media_model.Media, format string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

//	type MediaServiceImpl struct {
//		filesBuffer sync.Pool
//
//		typeService models.MediaTypeService
//		fileService models.FileService
//
//		AlbumService models.AlbumService
//		mediaMap     map[models.ContentId]*models.Media
//
//		streamerMap map[models.ContentId]*models.VideoStreamer
//
//		mediaCache *sturdyc.Client[[]byte]
//
//		collection *mongo.Collection
//
//		ollama *ollama.Client
//
//		doImageRecog bool
//
//		log *zerolog.Logger
//
//		mediaLock sync.RWMutex
//
//		streamerLock sync.RWMutex
//	}
var exifd *exiftool.Exiftool

type cacheKey string

const (
	CacheIdKey      cacheKey = "cacheId"
	CacheQualityKey cacheKey = "cacheQuality"
	CachePageKey    cacheKey = "cachePageNum"
	CacheMediaKey   cacheKey = "cacheMedia"

	HighresMaxSize = 2500
	ThumbMaxSize   = 1000
)

//	func init() {
//		var err error
//		exif, err = exiftool.NewExiftool(
//			exiftool.Api("largefilesupport"),
//			exiftool.Buffer([]byte{}, 1000*100),
//		)
//		if err != nil {
//			panic(err)
//		}
//
//		vips.LoggingSettings(nil, vips.LogLevelWarning)
//		vips.Startup(&vips.Config{})
//	}
//
// func NewMediaService(
//
//	fileService models.FileService, mediaTypeServ models.MediaTypeService, albumService models.AlbumService,
//	col *mongo.Collection, logger *zerolog.Logger,
//
//	) (*MediaServiceImpl, error) {
//		ms := &MediaServiceImpl{
//			mediaMap:     make(map[models.ContentId]*models.Media),
//			streamerMap:  make(map[models.ContentId]*models.VideoStreamer),
//			typeService:  mediaTypeServ,
//			mediaCache:   sturdyc.New[[]byte](1500, 10, time.Hour, 10),
//			fileService:  fileService,
//			collection:   col,
//			AlbumService: albumService,
//			filesBuffer:  sync.Pool{New: func() any { return &[]byte{} }},
//			log:          logger,
//			doImageRecog: os.Getenv("OLLAMA_HOST") != "",
//		}
//
//		client, err := ollama.ClientFromEnvironment()
//		if err != nil {
//			return nil, err
//		}
//
//		ms.ollama = client
//
//		indexModel := mongo.IndexModel{
//			Keys:    bson.D{{Key: "contentId", Value: 1}},
//			Options: (&options.IndexOptions{}).SetUnique(true),
//		}
//		_, err = col.Indexes().CreateOne(context.Background(), indexModel)
//		if err != nil {
//			return nil, err
//		}
//
//		ret, err := ms.collection.Find(context.Background(), bson.M{})
//		if err != nil {
//			return nil, errors.WithStack(err)
//		}
//
//		ms.mediaLock.Lock()
//		defer ms.mediaLock.Unlock()
//
//		cursorContext := context.Background()
//		for ret.Next(cursorContext) {
//			m := &models.Media{}
//			err = ret.Decode(m)
//			if err != nil {
//				return nil, errors.WithStack(err)
//			}
//			ms.mediaMap[m.ID()] = m
//		}
//
//		return ms, nil
//	}
//
//	func (ms *MediaServiceImpl) Size() int {
//		return len(ms.mediaMap)
//	}
//
//	func (ms *MediaServiceImpl) Add(m *models.Media) error {
//		if m == nil {
//			return errors.ErrMediaNil
//		}
//
//		if m.ID() == "" {
//			return errors.ErrMediaNoId
//		}
//
//		if m.GetPageCount() == 0 {
//			return errors.ErrMediaNoPages
//		}
//
//		if m.Width == 0 || m.Height == 0 {
//			ms.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Media %s has height %d and width %d", m.ID(), m.Height, m.Width) })
//			return errors.ErrMediaNoDimensions
//		}
//
//		if len(m.FileIDs) == 0 {
//			return errors.ErrMediaNoFiles
//		}
//
//		mt := ms.GetMediaType(m)
//		if mt.Mime == "" || mt.Mime == "generic" {
//			return errors.ErrMediaBadMime
//		}
//
//		isVideo := mt.Video
//		if isVideo && m.Duration == 0 {
//			return errors.ErrMediaNoDuration
//		}
//
//		if !isVideo && m.Duration != 0 {
//			return errors.ErrMediaHasDuration
//		}
//
//		ms.mediaLock.Lock()
//		defer ms.mediaLock.Unlock()
//
//		if ms.mediaMap[m.ID()] != nil {
//			return errors.ErrMediaAlreadyExists
//		}
//
//		if !m.IsImported() {
//			m.SetImported(true)
//			m.MediaID = primitive.NewObjectID()
//			_, err := ms.collection.InsertOne(context.Background(), m)
//			if err != nil {
//				return errors.WithStack(err)
//			}
//		}
//
//		ms.mediaMap[m.ID()] = m
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) TypeService() models.MediaTypeService {
//		return ms.typeService
//	}
//
//	func (ms *MediaServiceImpl) Get(mId models.ContentId) *models.Media {
//		if mId == "" {
//			return nil
//		}
//
//		ms.mediaLock.RLock()
//		defer ms.mediaLock.RUnlock()
//		m := ms.mediaMap[mId]
//
//		return m
//	}
//
//	func (ms *MediaServiceImpl) GetAll() []*models.Media {
//		ms.mediaLock.RLock()
//		defer ms.mediaLock.RUnlock()
//		medias := wl_slices.MapToValues(ms.mediaMap)
//		return wl_slices.Convert[*models.Media](medias)
//	}
//
//	func (ms *MediaServiceImpl) Del(cId models.ContentId) error {
//		m := ms.Get(cId)
//		err := ms.removeCacheFiles(m)
//		if err != nil && !errors.Is(err, errors.ErrNoCache) {
//			return err
//		}
//
//		err = ms.AlbumService.RemoveMediaFromAny(m.ID())
//		if err != nil {
//			return err
//		}
//
//		filter := bson.M{"contentId": m.ID()}
//		_, err = ms.collection.DeleteOne(context.Background(), filter)
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		ms.mediaLock.Lock()
//		defer ms.mediaLock.Unlock()
//		delete(ms.mediaMap, m.ID())
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) HideMedia(m *models.Media, hidden bool) error {
//		filter := bson.M{"contentId": m.ID()}
//		_, err := ms.collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"hidden": hidden}})
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		m.Hidden = hidden
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) FetchCacheImg(m *models.Media, q models.MediaQuality, pageNum int) ([]byte, error) {
//		cacheId := m.ID() + string(q) + strconv.Itoa(pageNum)
//
//		ctx := context.Background()
//		ctx = context.WithValue(ctx, CacheIdKey, cacheId)
//		ctx = context.WithValue(ctx, CacheQualityKey, q)
//		ctx = context.WithValue(ctx, CachePageKey, pageNum)
//		ctx = context.WithValue(ctx, CacheMediaKey, m)
//
//		cache, err := ms.mediaCache.GetOrFetch(ctx, cacheId, ms.getFetchMediaCacheImage)
//		if err != nil {
//			return nil, errors.WithStack(err)
//		}
//		return cache, nil
//	}
//
//	func (ms *MediaServiceImpl) StreamCacheVideo(m *models.Media, startByte, endByte int) ([]byte, error) {
//		return nil, errors.NotImplemented("StreamCacheVideo")
//		// cacheKey := fmt.Sprintf("%s-STREAM %d-%d", m.ID(), startByte, endByte)
//
//		// ctx := context.Background()
//		// ctx = context.WithValue(ctx, "cacheKey", cacheKey)
//		// ctx = context.WithValue(ctx, "startByte", startByte)
//		// ctx = context.WithValue(ctx, "endByte", endByte)
//		// ctx = context.WithValue(ctx, "Media", m)
//
//		// video, err := fetchAndCacheVideo(m.(*Media), startByte, endByte)
//		// if err != nil {
//		// 	return nil, err
//		// }
//		// cache, err := mr.mediaCache.GetFetch(ctx, cacheKey, fetchAndCacheVideo)
//		// if err != nil {
//		// 	return nil, err
//		// }
//		// return cache, nil
//	}
//
//	type justContentId struct {
//		Cid string `bson:"contentId"`
//	}
//
// func (ms *MediaServiceImpl) AdjustMediaDates(
//
//	anchor *models.Media, newTime time.Time, extraMedias []*models.Media,
//
//	) error {
//		offset := newTime.Sub(anchor.GetCreateDate())
//
//		anchor.SetCreateDate(anchor.GetCreateDate().Add(offset))
//
//		for _, m := range extraMedias {
//			m.SetCreateDate(m.GetCreateDate().Add(offset))
//		}
//
//		// TODO - update media date in DB
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) IsCached(m *models.Media) bool {
//		cacheFile, err := ms.getCacheFile(m, models.LowRes, 0)
//		return cacheFile != nil && err == nil
//	}
//
//	func (ms *MediaServiceImpl) IsFileDisplayable(f *fileTree.WeblensFileImpl) bool {
//		ext := filepath.Ext(f.Filename())
//		return ms.typeService.ParseExtension(ext).Displayable
//	}
//
//	func (ms *MediaServiceImpl) AddFileToMedia(m *models.Media, f *fileTree.WeblensFileImpl) error {
//		if slices.ContainsFunc(
//			m.FileIDs, func(fId fileTree.FileId) bool {
//				return fId == f.ID()
//			},
//		) {
//			return nil
//		}
//
//		filter := bson.M{"contentId": m.ID()}
//		update := bson.M{"$addToSet": bson.M{"fileIds": f.ID()}}
//		_, err := ms.collection.UpdateOne(context.Background(), filter, update)
//		if err != nil {
//			return err
//		}
//
//		m.AddFile(f)
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {
//		filter := bson.M{"contentId": media.ID()}
//		update := bson.M{"$pull": bson.M{"fileIds": fileId}}
//		_, err := ms.collection.UpdateOne(context.Background(), filter, update)
//		if err != nil {
//			return err
//		}
//
//		media.RemoveFile(fileId)
//
//		if len(media.FileIDs) == 1 && media.FileIDs[0] == fileId {
//			return ms.Del(media.ID())
//		}
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) Cleanup() error {
//		for _, m := range ms.mediaMap {
//			fs, missing, err := ms.fileService.GetFiles(m.FileIDs)
//			if err != nil {
//				return err
//			}
//			for _, f := range fs {
//				if f.GetPortablePath().RootName() != "USERS" {
//					missing = append(missing, f.ID())
//				}
//			}
//
//			for _, fId := range missing {
//				err = ms.RemoveFileFromMedia(m, fId)
//				if err != nil {
//					return err
//				}
//			}
//		}
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) Drop() error {
//		ms.mediaLock.Lock()
//		defer ms.mediaLock.Unlock()
//
//		// Drop media collection in mongo
//		err := ms.collection.Drop(context.Background())
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		cacheTree, err := ms.fileService.GetFileTreeByName(CachesTreeKey)
//		if err != nil {
//			return err
//		}
//
//		thumbsDir, err := cacheTree.GetRoot().GetChild(ThumbsDirName)
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		// Delete all cache files on disk
//		thumbs := thumbsDir.GetChildren()
//		for _, thumb := range thumbs {
//			err = ms.fileService.DeleteCacheFile(thumb)
//			if err != nil {
//				return err
//			}
//		}
//
//		// Evict all keys from cache
//		for _, cacheKey := range ms.mediaCache.ScanKeys() {
//			ms.mediaCache.Delete(cacheKey)
//		}
//
//		ms.mediaMap = map[models.ContentId]*models.Media{}
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) GetProminentColors(media *models.Media) (prom []string, err error) {
//		var i image.Image
//		thumbBytes, err := ms.FetchCacheImg(media, models.LowRes, 0)
//		if err != nil {
//			return
//		}
//
//		i, err = webp.Decode(bytes.NewBuffer(thumbBytes))
//		if err != nil {
//			return
//		}
//
//		promColors, err := prominentcolor.Kmeans(i)
//		prom = wl_slices.Map(promColors, func(p prominentcolor.ColorItem) string { return p.AsString() })
//		return
//	}
//
// func (ms *MediaServiceImpl) StreamVideo(
//
//	m *models.Media, u *models.User, share *models.FileShare,
//
//	) (*models.VideoStreamer, error) {
//		if !ms.GetMediaType(m).Video {
//			return nil, errors.WithStack(errors.ErrMediaNotVideo)
//		}
//
//		ms.streamerLock.Lock()
//		defer ms.streamerLock.Unlock()
//
//		var streamer *models.VideoStreamer
//		var ok bool
//		if streamer, ok = ms.streamerMap[m.ID()]; !ok {
//			f, err := ms.fileService.GetFileByContentId(m.ContentID)
//			if err != nil {
//				return nil, err
//			}
//
//			thumbs, err := ms.fileService.GetThumbsDir()
//			if err != nil {
//				return nil, err
//			}
//			streamer = models.NewVideoStreamer(f, thumbs.AbsPath())
//			ms.streamerMap[m.ID()] = streamer
//		}
//
//		return streamer, nil
//	}
//
//	func (ms *MediaServiceImpl) SetMediaLiked(mediaId models.ContentId, liked bool, username string) error {
//		m := ms.Get(mediaId)
//		if m == nil {
//			return errors.Errorf("Could not find media with id [%s] while trying to update liked array", mediaId)
//		}
//
//		filter := bson.M{"contentId": mediaId}
//		var update bson.M
//		if liked && len(m.LikedBy) == 0 {
//			update = bson.M{"$set": bson.M{"likedBy": []string{username}}}
//		} else if liked && len(m.LikedBy) == 0 {
//			update = bson.M{"$addToSet": bson.M{"likedBy": username}}
//		} else {
//			update = bson.M{"$pull": bson.M{"likedBy": username}}
//		}
//
//		_, err := ms.collection.UpdateOne(context.Background(), filter, update)
//		if err != nil {
//			return err
//		}
//
//		if liked {
//			m.LikedBy = wl_slices.AddToSet(m.LikedBy, username)
//		} else {
//			m.LikedBy = wl_slices.Filter(
//				m.LikedBy, func(u string) bool {
//					return u != username
//				},
//			)
//		}
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) GetMediaConverted(m *models.Media, format string) ([]byte, error) {
//		f, err := ms.fileService.GetFileByTree(m.FileIDs[0], UsersTreeKey)
//		if err != nil {
//			return nil, err
//		}
//
//		img, err := ms.loadImageFromFile(f, ms.GetMediaType(m))
//		if err != nil {
//			return nil, err
//		}
//
//		var blob []byte
//		switch format {
//		case "png":
//			blob, _, err = img.ExportPng(nil)
//		case "jpeg":
//			blob, _, err = img.ExportJpeg(nil)
//		default:
//			return nil, errors.Errorf("Unknown media convert format [%s]", format)
//		}
//		return blob, err
//	}
//
//	func (ms *MediaServiceImpl) removeCacheFiles(media *models.Media) error {
//		thumbCache, err := ms.getCacheFile(media, models.LowRes, 0)
//		if err != nil && !errors.Is(err, errors.ErrNoFile) {
//			return err
//		}
//
//		if thumbCache != nil {
//			err = ms.fileService.DeleteCacheFile(thumbCache)
//			if err != nil {
//				return err
//			}
//		}
//
//		highresCacheFile, err := ms.getCacheFile(media, models.HighRes, 0)
//		if err != nil && !errors.Is(err, errors.ErrNoFile) {
//			return err
//		}
//
//		if highresCacheFile != nil {
//			err = ms.fileService.DeleteCacheFile(highresCacheFile)
//			if err != nil {
//				return err
//			}
//		}
//
//		return nil
//	}
func NewMediaFromFile(ctx *context.AppContext, file *file_model.WeblensFileImpl) (m *media_model.Media, err error) {
	m, err = media_model.GetMediaById(ctx, file.GetContentId())
	if err != nil {
		m = &media_model.Media{}
	}

	fileMetas := exifd.ExtractMetadata(file.GetPortablePath().ToAbsolute())

	for _, fileMeta := range fileMetas {
		if fileMeta.Err != nil {
			return nil, fileMeta.Err
		}
	}

	if m.CreateDate.Unix() <= 0 {
		createDate, err := getCreateDateFromExif(fileMetas[0].Fields, file)
		if err != nil {
			return nil, err
		}
		m.CreateDate = createDate
	}

	if m.MimeType == "" {
		ext := file.GetPortablePath().Ext()
		mType := media_model.ParseExtension(ext)
		m.MimeType = mType.Mime

		if media_model.ParseMime(m.MimeType).Video {
			m.Width = int(fileMetas[0].Fields["ImageWidth"].(float64))
			m.Height = int(fileMetas[0].Fields["ImageHeight"].(float64))

			duration, err := getVideoDurationMs(file.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}
			m.Duration = duration
		}
	}

	mType := GetMediaType(m)
	if !mType.IsSupported() {
		return nil, media_model.ErrMediaBadMimeType
	}

	if mType.IsMultiPage() {
		m.PageCount = int(fileMetas[0].Fields["PageCount"].(float64))
	} else {
		m.PageCount = 1
	}

	_, err = handleCacheCreation(ctx, m, file)
	if err != nil {
		return
	}

	// if !mType.Video && ms.doImageRecog {
	// 	go func() {
	// 		err := ms.GetImageTags(m, thumb)
	// 		if err != nil {
	// 			ms.log.Error().Stack().Err(err).Msg("")
	// 		}
	// 	}()
	// }

	return
}

func GetMediaType(m *media_model.Media) media_model.MediaType {
	return media_model.ParseMime(m.MimeType)
}

//	func (ms *MediaServiceImpl) GetMediaTypes() models.MediaTypeService {
//		return ms.typeService
//	}
//
//	func (ms *MediaServiceImpl) RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []*models.Media {
//		var medias []*models.Media
//
//		for _, f := range folders {
//			if f == nil {
//				ms.log.Warn().Msg("Skipping recursive media lookup for non-existent folder")
//				continue
//			}
//			if !f.IsDir() {
//				if ms.IsFileDisplayable(f) {
//					m := ms.Get(f.GetContentId())
//					if m != nil {
//						medias = append(medias, m)
//					}
//				}
//				continue
//			}
//			err := f.RecursiveMap(
//				func(f *fileTree.WeblensFileImpl) error {
//					if !f.IsDir() && ms.IsFileDisplayable(f) {
//						m := ms.Get(f.GetContentId())
//						if m != nil {
//							medias = append(medias, m)
//						}
//					}
//					return nil
//				},
//			)
//			if err != nil {
//				ms.log.Error().Stack().Err(err).Msg("")
//			}
//		}
//
//		return medias
//	}
func handleCacheCreation(ctx *context.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) (thumbBytes []byte, err error) {
	mType := GetMediaType(m)

	if !mType.Video {
		img, err := loadImageFromFile(file, mType)
		if err != nil {
			return nil, err
		}

		m.PageCount = img.Pages()
		// Read image dimensions
		m.Height = img.Height()
		m.Width = img.Width()

		if mType.IsMultiPage() {
			fullPdf, err := file.ReadAll()
			if err != nil {
				return nil, errors.WithStack(err)
			}
			for page := range m.PageCount {
				vipsPage := vips.IntParameter{}
				vipsPage.Set(page)
				img, err := vips.LoadImageFromBuffer(fullPdf, &vips.ImportParams{Page: vipsPage})
				if err != nil {
					return nil, errors.WithStack(err)
				}

				err = handleNewHighRes(ctx, m, img, page)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err = handleNewHighRes(ctx, m, img, 0)
			if err != nil {
				return nil, err
			}
		}

		// Resize thumb image if too big
		if m.Width > ThumbMaxSize || m.Height > ThumbMaxSize {
			var thumbWidth, thumbHeight uint
			if m.Width > m.Height {
				thumbWidth = ThumbMaxSize
				thumbHeight = uint(float64(ThumbMaxSize) / float64(m.Width) * float64(m.Height))
			} else {
				thumbHeight = ThumbMaxSize
				thumbWidth = uint(float64(ThumbMaxSize) / float64(m.Height) * float64(m.Width))
			}
			ctx.Logger.Trace().Func(func(e *zerolog.Event) {
				e.Msgf("Resizing %s thumb image to %dx%d", file.GetPortablePath(), thumbWidth, thumbHeight)
			})
			err = img.Resize(float64(thumbHeight)/float64(m.Height), vips.KernelAuto)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		// Create and write thumb cache file
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, errors.WithStack(err)
		} else if err == nil {
			blob, _, err := img.ExportWebp(nil)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			_, err = thumb.Write(blob)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			m.SetLowresCacheFile(thumb)

			thumbBytes = blob
		}

	} else {
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, errors.WithStack(err)
		} else if err == nil {
			thumbBytes, err = generateVideoThumbnail(file.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}
			_, err = thumb.Write(thumbBytes)
			if err != nil {
				return nil, err
			}
			m.SetLowresCacheFile(thumb)
		}

	}

	return thumbBytes, nil
}

func handleNewHighRes(ctx *context.AppContext, m *media_model.Media, img *vips.ImageRef, page int) error {
	// Resize highres image if too big
	if m.Width > HighresMaxSize || m.Height > HighresMaxSize {
		var fullHeight int
		if m.Width > m.Height {
			// fullWidth = HighresMaxSize
			fullHeight = HighresMaxSize * m.Height / m.Width
		} else {
			fullHeight = HighresMaxSize
			// fullWidth = HighresMaxSize * m.Width / m.Height
		}

		err := img.Resize(float64(fullHeight)/float64(m.Height), vips.KernelAuto)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// Create and write highres cache file
	highres, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.HighRes), page)
	if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
		return errors.WithStack(err)
	} else if err == nil {
		blob, _, err := img.ExportWebp(nil)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = highres.Write(blob)
		if err != nil {
			return errors.WithStack(err)
		}
		m.SetHighresCacheFiles(highres, page)
	}

	return nil
}

//	func (ms *MediaServiceImpl) getFetchMediaCacheImage(ctx context.Context) (data []byte, err error) {
//		defer internal.RecoverPanic("Fetching media image had panic")
//
//		m := ctx.Value(CacheMediaKey).(*models.Media)
//		q := ctx.Value(CacheQualityKey).(models.MediaQuality)
//		pageNum, _ := ctx.Value(CachePageKey).(int)
//
//		f, err := ms.getCacheFile(m, q, pageNum)
//		if err != nil {
//			return nil, err
//		}
//
//		if f == nil {
//			return nil, errors.Errorf("This should never happen... file is nil in GetFetchMediaCacheImage")
//		}
//
//		ms.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Reading image cache for media [%s]", m.ID()) })
//
//		data, err = f.ReadAll()
//		if err != nil {
//			return nil, err
//		}
//		if len(data) == 0 {
//			err = errors.Errorf("displayable bytes empty")
//			return nil, err
//		}
//
//		return data, nil
//	}
//
// func (ms *MediaServiceImpl) getCacheFile(
//
//	m *models.Media, quality models.MediaQuality, pageNum int,
//
//	) (*fileTree.WeblensFileImpl, error) {
//		if quality == models.LowRes && m.GetLowresCacheFile() != nil {
//			return m.GetLowresCacheFile(), nil
//		} else if quality == models.HighRes && m.GetHighresCacheFiles(pageNum) != nil {
//			return m.GetHighresCacheFiles(pageNum), nil
//		}
//
//		filename := m.FmtCacheFileName(quality, pageNum)
//		cacheFile, err := ms.fileService.GetMediaCacheByFilename(filename)
//		if err != nil {
//			return nil, errors.WithStack(errors.ErrNoCache)
//		}
//
//		if quality == models.LowRes {
//			m.SetLowresCacheFile(cacheFile)
//		} else if quality == models.HighRes {
//			m.SetHighresCacheFiles(cacheFile, pageNum)
//		} else {
//			return nil, errors.Errorf("Unknown media quality [%s]", quality)
//		}
//
//		return cacheFile, nil
//	}
func loadImageFromFile(f *file_model.WeblensFileImpl, mType media_model.MediaType) (*vips.ImageRef, error) {
	filePath := f.GetPortablePath().ToAbsolute()
	var img *vips.ImageRef
	var err error

	// Sony RAWs do not play nice with govips. Should fall back to imagick but it thinks its a TIFF.
	// The real libvips figures this out, adding an intermediary step using dcraw to convert to a real TIFF
	// and continuing processing from there solves this issue, and is surprisingly fast. Everyone say "Thank you dcraw"
	if strings.HasSuffix(filePath, "ARW") || strings.HasSuffix(filePath, "CR2") {
		cmd := exec.Command("dcraw", "-T", "-w", "-h", "-c", filePath)
		var stdb, errb bytes.Buffer
		cmd.Stderr = &errb
		cmd.Stdout = &stdb

		err = cmd.Run()
		if err != nil {
			return nil, errors.WithStack(errors.New(err.Error() + "\n" + errb.String()))
		}

		img, err = vips.NewImageFromReader(&stdb)
	} else {
		img, err = vips.NewImageFromFile(filePath)
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	// PDFs and HEIFs do not need to be rotated.
	if !mType.IsMultiPage() && !mType.IsMime("image/heif") {
		// Rotate image based on exif data
		err = img.AutoRotate()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return img, nil
}

// var recogLock sync.Mutex
//
//	func (ms *MediaServiceImpl) GetImageTags(m *models.Media, imageBytes []byte) error {
//		if !ms.doImageRecog {
//			return nil
//		}
//
//		recogLock.Lock()
//		defer recogLock.Unlock()
//		img, err := vips.NewImageFromBuffer(imageBytes)
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		blob, _, err := img.ExportJpeg(nil)
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		stream := false
//		req := &ollama.GenerateRequest{
//			Model:  "llava:13b",
//			Prompt: "describe this image using a list of single words seperated only by commas. do not include any text other than these words",
//			Images: []ollama.ImageData{blob},
//			Stream: &stream,
//			Options: map[string]any{
//				"n_ctx":       1024,
//				"num_predict": 25,
//			},
//		}
//
//		tagsString := ""
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
//		defer cancel()
//		doneChan := make(chan struct{})
//		err = ms.ollama.Generate(ctx, req, func(resp ollama.GenerateResponse) error {
//			ms.log.Trace().Msgf("Got recognition response %s", resp.Response)
//			tagsString = resp.Response
//
//			if resp.Done {
//				close(doneChan)
//			}
//
//			return nil
//		})
//
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		select {
//		case <-doneChan:
//		case <-ctx.Done():
//		}
//
//		if ctx.Err() != nil {
//			return errors.WithStack(ctx.Err())
//		}
//
//		tags := strings.Split(tagsString, ",")
//		for i, tag := range tags {
//			tags[i] = strings.ToLower(strings.ReplaceAll(tag, " ", ""))
//		}
//
//		_, err = ms.collection.UpdateOne(context.Background(), bson.M{"contentId": m.ID()}, bson.M{"$set": bson.M{"recognitionTags": tags}})
//		if err != nil {
//			return err
//		}
//		m.SetRecognitionTags(tags)
//
//		return nil
//	}
func getCreateDateFromExif(exif map[string]any, file *file_model.WeblensFileImpl) (createDate time.Time, err error) {
	r, ok := exif["SubSecCreateDate"]
	if !ok {
		r, ok = exif["MediaCreateDate"]
	}
	if ok {
		createDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05.00-07:00", r.(string))
		}
		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
		}
		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05-07:00", r.(string))
		}
		if err != nil {
			createDate = file.ModTime()
		}
	} else {
		createDate = file.ModTime()
	}

	return createDate, nil
}

func generateVideoThumbnail(filepath string) ([]byte, error) {
	const frameNum = 10

	buf := bytes.NewBuffer(nil)
	errOut := bytes.NewBuffer(nil)

	// Get the 10th frame of the video and save it to the cache as the thumbnail
	// "Highres" for video is the video itself
	err := ffmpeg.Input(filepath).Filter(
		"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
	).Output(
		"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
	).WithOutput(buf).WithErrorOutput(errOut).Run()
	if err != nil {
		return nil, errors.WithStack(errors.New(err.Error() + errOut.String()))
	}

	return buf.Bytes(), nil
}

func getVideoDurationMs(filepath string) (int, error) {
	probeJson, err := ffmpeg.Probe(filepath)
	if err != nil {
		return 0, err
	}
	probeResult := map[string]any{}
	err = json.Unmarshal([]byte(probeJson), &probeResult)
	if err != nil {
		return 0, err
	}

	formatChunk, ok := probeResult["format"].(map[string]any)
	if !ok {
		return 0, errors.Errorf("invalid movie format")
	}
	duration, err := strconv.ParseFloat(formatChunk["duration"].(string), 32)
	if err != nil {
		return 0, err
	}
	return int(duration) * 1000, nil
}
