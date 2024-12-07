package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethanrous/bimg"
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/image/webp"

	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/gographics/imagick.v3/imagick"
)

var _ models.MediaService = (*MediaServiceImpl)(nil)

type MediaServiceImpl struct {
	filesBuffer sync.Pool

	typeService models.MediaTypeService
	fileService models.FileService

	AlbumService models.AlbumService
	mediaMap     map[models.ContentId]*models.Media

	streamerMap map[models.ContentId]*models.VideoStreamer

	mediaCache *sturdyc.Client[[]byte]

	collection *mongo.Collection

	mediaLock sync.RWMutex

	streamerLock sync.RWMutex
}

var exif *exiftool.Exiftool

type cacheKey string

const (
	CacheIdKey      cacheKey = "cacheId"
	CacheQualityKey cacheKey = "cacheQuality"
	CachePageKey    cacheKey = "cachePageNum"
	CacheMediaKey   cacheKey = "cacheMedia"
)

func init() {
	var err error
	exif, err = exiftool.NewExiftool(
		exiftool.Api("largefilesupport"),
		exiftool.Buffer([]byte{}, 1000*100),
	)
	if err != nil {
		panic(err)
	}

	imagick.Initialize()
}

func NewMediaService(
	fileService models.FileService, mediaTypeServ models.MediaTypeService, albumService models.AlbumService,
	col *mongo.Collection,
) (*MediaServiceImpl, error) {
	ms := &MediaServiceImpl{
		mediaMap:     make(map[models.ContentId]*models.Media),
		streamerMap:  make(map[models.ContentId]*models.VideoStreamer),
		typeService:  mediaTypeServ,
		mediaCache:   sturdyc.New[[]byte](1500, 10, time.Hour, 10),
		fileService:  fileService,
		collection:   col,
		AlbumService: albumService,
		filesBuffer:  sync.Pool{New: func() any { return &[]byte{} }},
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "contentId", Value: 1}},
		Options: (&options.IndexOptions{}).SetUnique(true),
	}
	_, err := col.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, err
	}

	ret, err := ms.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, werror.WithStack(err)
	}

	ms.mediaLock.Lock()
	defer ms.mediaLock.Unlock()

	cursorContext := context.Background()
	for ret.Next(cursorContext) {
		m := &models.Media{}
		err = ret.Decode(m)
		if err != nil {
			return nil, werror.WithStack(err)
		}
		ms.mediaMap[m.ID()] = m
	}

	return ms, nil
}

func (ms *MediaServiceImpl) Size() int {
	return len(ms.mediaMap)
}

func (ms *MediaServiceImpl) Add(m *models.Media) error {
	if m == nil {
		return werror.ErrMediaNil
	}

	if m.ID() == "" {
		return werror.ErrMediaNoId
	}

	if m.GetPageCount() == 0 {
		return werror.ErrMediaNoPages
	}

	if m.Width == 0 || m.Height == 0 {
		log.Debug.Printf("Media %s has height %d and width %d", m.ID(), m.Height, m.Width)
		return werror.ErrMediaNoDimensions
	}

	if len(m.FileIDs) == 0 {
		return werror.ErrMediaNoFiles
	}

	mt := ms.GetMediaType(m)
	if mt.Mime == "" || mt.Mime == "generic" {
		return werror.ErrMediaBadMime
	}

	isVideo := mt.IsVideo()
	if isVideo && m.Duration == 0 {
		return werror.ErrMediaNoDuration
	}

	if !isVideo && m.Duration != 0 {
		return werror.ErrMediaHasDuration
	}

	ms.mediaLock.Lock()
	defer ms.mediaLock.Unlock()

	if ms.mediaMap[m.ID()] != nil {
		return werror.ErrMediaAlreadyExists
	}

	if !m.IsImported() {
		m.SetImported(true)
		m.MediaID = primitive.NewObjectID()
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

	ms.mediaLock.RLock()
	m := ms.mediaMap[mId]
	ms.mediaLock.RUnlock()

	if reflect2.IsNil(m) {
		return nil
	}

	return m
}

func (ms *MediaServiceImpl) GetAll() []*models.Media {
	ms.mediaLock.RLock()
	defer ms.mediaLock.RUnlock()
	medias := internal.MapToValues(ms.mediaMap)
	return internal.SliceConvert[*models.Media](medias)
}

func (ms *MediaServiceImpl) Del(cId models.ContentId) error {
	m := ms.Get(cId)
	err := ms.removeCacheFiles(m)
	if err != nil && !errors.Is(err, werror.ErrNoCache) {
		return err
	}

	err = ms.AlbumService.RemoveMediaFromAny(m.ID())
	if err != nil {
		return err
	}

	filter := bson.M{"contentId": m.ID()}
	_, err = ms.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return werror.WithStack(err)
	}

	ms.mediaLock.Lock()
	defer ms.mediaLock.Unlock()
	delete(ms.mediaMap, m.ID())

	return nil
}

func (ms *MediaServiceImpl) HideMedia(m *models.Media, hidden bool) error {
	filter := bson.M{"contentId": m.ID()}
	_, err := ms.collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"hidden": hidden}})
	if err != nil {
		return werror.WithStack(err)
	}

	m.Hidden = hidden

	return nil
}

func (ms *MediaServiceImpl) FetchCacheImg(m *models.Media, q models.MediaQuality, pageNum int) ([]byte, error) {
	cacheId := m.ID() + string(q) + strconv.Itoa(pageNum)

	ctx := context.Background()
	ctx = context.WithValue(ctx, CacheIdKey, cacheId)
	ctx = context.WithValue(ctx, CacheQualityKey, q)
	ctx = context.WithValue(ctx, CachePageKey, pageNum)
	ctx = context.WithValue(ctx, CacheMediaKey, m)

	cache, err := ms.mediaCache.GetFetch(ctx, cacheId, ms.getFetchMediaCacheImage)
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

type justContentId struct {
	Cid string `bson:"contentId"`
}

func (ms *MediaServiceImpl) GetFilteredMedia(
	requester *models.User, sort string, sortDirection int, excludeIds []models.ContentId,
	allowRaw bool, allowHidden bool,
) ([]*models.Media, error) {
	slices.Sort(excludeIds)

	pipe := bson.A{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "owner", Value: requester.GetUsername()},
				{Key: "fileIds", Value: bson.D{
					{Key: "$exists", Value: true}, {Key: "$ne", Value: bson.A{}},
				}}},
			},
		},
	}

	if !allowHidden {
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.D{{Key: "hidden", Value: false}}}})
	}

	pipe = append(pipe, bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: sortDirection}}}})
	pipe = append(pipe, bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: false}, {Key: "contentId", Value: true}}}})
	log.Debug.Printf("Filtering media with pipe: %v", pipe)

	cur, err := ms.collection.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	allIds := []justContentId{}
	err = cur.All(context.Background(), &allIds)
	if err != nil {
		return nil, werror.WithStack(err)
	}
	log.Debug.Printf("Found %d media", len(allIds))

	medias := make([]*models.Media, 0, len(allIds))
	for _, id := range allIds {
		log.Debug.Println("Getting media with id", id.Cid)
		m := ms.Get(id.Cid)
		if m != nil {
			if m.MimeType == "application/pdf" {
				continue
			}
			if !allowRaw {
				mt := ms.GetMediaType(m)
				if mt.IsRaw() {
					continue
				}
			}
			medias = append(medias, m)
		}
	}

	// ms.mediaLock.RLock()
	// allMs := internal.MapToValues(ms.mediaMap)
	// ms.mediaLock.RUnlock()
	// allMs = internal.Filter(
	// 	allMs, func(m *models.Media) bool {
	// 		mt := ms.GetMediaType(m)
	// 		if mt.Mime == "" || (mt.IsRaw() && !allowRaw) || (m.IsHidden() && !allowHidden) || m.GetOwner() != requester.GetUsername() || len(m.GetFiles()) == 0 || mt.IsMime("application/pdf") {
	// 			return false
	// 		}
	//
	// 		// Exclude Media if it is present in the filter
	// 		_, e := slices.BinarySearch(excludeIds, m.ID())
	// 		return !e
	// 	},
	// )
	//
	// slices.SortFunc(
	// 	allMs, func(a, b *models.Media) int { return b.GetCreateDate().Compare(a.GetCreateDate()) * sortDirection },
	// )

	return medias, nil
}

func (ms *MediaServiceImpl) AdjustMediaDates(
	anchor *models.Media, newTime time.Time, extraMedias []*models.Media,
) error {
	offset := newTime.Sub(anchor.GetCreateDate())

	anchor.SetCreateDate(anchor.GetCreateDate().Add(offset))

	for _, m := range extraMedias {
		m.SetCreateDate(m.GetCreateDate().Add(offset))
	}

	// TODO - update media date in DB

	return nil
}

func (ms *MediaServiceImpl) IsCached(m *models.Media) bool {
	cacheFile, err := ms.getCacheFile(m, models.LowRes, 0)
	return cacheFile != nil && err == nil
}

func (ms *MediaServiceImpl) IsFileDisplayable(f *fileTree.WeblensFileImpl) bool {
	ext := filepath.Ext(f.Filename())
	return ms.typeService.ParseExtension(ext).Displayable
}

func (ms *MediaServiceImpl) AddFileToMedia(m *models.Media, f *fileTree.WeblensFileImpl) error {
	if slices.ContainsFunc(
		m.FileIDs, func(fId fileTree.FileId) bool {
			return fId == f.ID()
		},
	) {
		return nil
	}

	filter := bson.M{"contentId": m.ID()}
	update := bson.M{"$addToSet": bson.M{"fileIds": f.ID()}}
	_, err := ms.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	m.AddFile(f)

	return nil
}

func (ms *MediaServiceImpl) RemoveFileFromMedia(media *models.Media, fileId fileTree.FileId) error {
	filter := bson.M{"contentId": media.ID()}
	update := bson.M{"$pull": bson.M{"fileIds": fileId}}
	_, err := ms.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	media.FileIDs = internal.Filter(
		media.FileIDs, func(fId fileTree.FileId) bool {
			return fId != fileId
		},
	)

	if len(media.FileIDs) == 1 && media.FileIDs[0] == fileId {
		return ms.Del(media.ID())
	}

	return nil
}

func (ms *MediaServiceImpl) Cleanup() error {
	for _, m := range ms.mediaMap {
		fs, missing, err := ms.fileService.GetFiles(m.FileIDs)
		if err != nil {
			return err
		}
		for _, f := range fs {
			if f.GetPortablePath().RootName() != "USERS" {
				missing = append(missing, f.ID())
			}
		}

		for _, fId := range missing {
			err = ms.RemoveFileFromMedia(m, fId)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	// 	err := types.SERV.FileTree.DeleteApiKey(child.ID())
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

	// return nil
}

func (ms *MediaServiceImpl) StreamVideo(
	m *models.Media, u *models.User, share *models.FileShare,
) (*models.VideoStreamer, error) {
	if !ms.GetMediaType(m).IsVideo() {
		return nil, werror.WithStack(werror.ErrMediaNotVideo)
	}

	ms.streamerLock.Lock()
	defer ms.streamerLock.Unlock()

	var streamer *models.VideoStreamer
	var ok bool
	if streamer, ok = ms.streamerMap[m.ID()]; !ok {
		f, err := ms.fileService.GetFileByContentId(m.ContentID)
		if err != nil {
			return nil, err
		}

		thumbs, err := ms.fileService.GetThumbsDir()
		if err != nil {
			return nil, err
		}
		streamer = models.NewVideoStreamer(f, thumbs.AbsPath())
		ms.streamerMap[m.ID()] = streamer
	}

	return streamer, nil
}

func (ms *MediaServiceImpl) SetMediaLiked(mediaId models.ContentId, liked bool, username models.Username) error {
	ms.mediaLock.Lock()
	defer ms.mediaLock.Unlock()
	m, ok := ms.mediaMap[mediaId]
	if !ok {
		return werror.Errorf("Could not find media with id [%s] while trying to update liked array", mediaId)
	}

	filter := bson.M{"contentId": mediaId}
	var update bson.M
	if liked && len(m.LikedBy) == 0 {
		update = bson.M{"$set": bson.M{"likedBy": []models.Username{username}}}
	} else if liked && len(m.LikedBy) == 0 {
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

func (ms *MediaServiceImpl) removeCacheFiles(media *models.Media) error {
	thumbCache, err := ms.getCacheFile(media, models.LowRes, 0)
	if err != nil && !errors.Is(err, werror.ErrNoFile) {
		return err
	}

	if thumbCache != nil {
		err = ms.fileService.DeleteCacheFile(thumbCache)
		if err != nil {
			return err
		}
	}

	highresCacheFile, err := ms.getCacheFile(media, models.HighRes, 0)
	if err != nil && !errors.Is(err, werror.ErrNoFile) {
		return err
	}

	if highresCacheFile != nil {
		err = ms.fileService.DeleteCacheFile(highresCacheFile)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	HighresSize = 2500
	ThumbSize   = 500
)

func (ms *MediaServiceImpl) LoadMediaFromFile(m *models.Media, file *fileTree.WeblensFileImpl) error {
	sw := internal.NewStopwatch("LoadMediaFromFile: " + file.Filename())
	fileMetas := exif.ExtractMetadata(file.AbsPath())
	sw.Lap("ExtractMetadata")

	for _, fileMeta := range fileMetas {
		if fileMeta.Err != nil {
			return fileMeta.Err
		}
	}
	sw.Lap("Read errors")

	var err error
	if m.CreateDate.Unix() <= 0 {
		r, ok := fileMetas[0].Fields["SubSecCreateDate"]
		if !ok {
			r, ok = fileMetas[0].Fields["MediaCreateDate"]
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
	sw.Lap("Read time")

	if m.MimeType == "" {
		mimeType, ok := fileMetas[0].Fields["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MimeType = mimeType

		if ms.typeService.ParseMime(m.MimeType).IsVideo() {
			probeJson, err := ffmpeg.Probe(file.AbsPath())
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
			duration, err := strconv.ParseFloat(formatChunk["duration"].(string), 32)
			if err != nil {
				return err
			}
			m.Duration = int(duration * 1000)
		}
	}
	sw.Lap("Read mime and duration")

	mType := ms.GetMediaType(m)
	if !mType.IsSupported() {
		return werror.ErrMediaBadMime
	}
	sw.Lap("Get MediaType")

	if mType.IsMultiPage() {
		m.PageCount = int(fileMetas[0].Fields["PageCount"].(float64))
	} else {
		m.PageCount = 1
	}
	sw.Lap("Count pages")

	if m.Rotate == "" {
		rotate := fileMetas[0].Fields["Orientation"]
		if rotate != nil {
			m.Rotate = rotate.(string)
		}
	}
	sw.Lap("Rotate")

	buf := *ms.filesBuffer.Get().(*[]byte)
	log.Trace.Func(func(l log.Logger) {
		if len(buf) > 0 {
			l.Printf("Re-using buffer of %d bytes", len(buf))
		}
	})
	sw.Lap("Get Buffer")

	if !mType.IsVideo() && !mType.IsMultiPage() {
		mw := imagick.NewMagickWand()
		defer mw.Destroy()
		sw.Lap("New MagickWand")

		err = mw.SetCompressionQuality(99)
		if err != nil {
			return werror.WithStack(err)
		}

		err = mw.ReadImage(file.AbsPath())
		if err != nil {
			return werror.WithStack(err)
		}
		sw.Lap("Read image")
		err = mw.AutoOrientImage()
		if err != nil {
			return werror.WithStack(err)
		}
		sw.Lap("Image orientation")

		width := mw.GetImageWidth()
		height := mw.GetImageHeight()

		m.Height = int(height)
		m.Width = int(width)
		sw.Lap("Image size")
		log.Trace.Printf("%s Image size: %dx%d", file.Filename(), width, height)

		err = mw.SetImageFormat("webp")
		if err != nil {
			return werror.WithStack(err)
		}
		sw.Lap("Image convert to webp")

		if width > HighresSize || height > HighresSize {
			var fullWidth, fullHeight uint
			if width > height {
				fullWidth = HighresSize
				fullHeight = HighresSize * uint(height) / uint(width)
			} else {
				fullHeight = HighresSize
				fullWidth = HighresSize * uint(width) / uint(height)
			}
			log.Trace.Printf("Resizing %s highres image to %dx%d", file.Filename(), fullWidth, fullHeight)

			err = mw.ScaleImage(fullWidth, fullHeight)
			if err != nil {
				return werror.WithStack(err)
			}
			sw.Lap("Image resize for fullres")
		}

		highres, err := ms.fileService.NewCacheFile(m, models.HighRes, 0)
		if err != nil && !errors.Is(err, werror.ErrFileAlreadyExists) {
			return werror.WithStack(err)
		} else if err == nil {
			blob, err := mw.GetImageBlob()
			if err != nil {
				return werror.WithStack(err)
			}
			_, err = highres.Write(blob)
			if err != nil {
				return werror.WithStack(err)
			}
			m.SetHighresCacheFiles(highres, 0)
			sw.Lap("Write highres cache file")
		}

		if width > ThumbSize || height > ThumbSize {
			var thumbWidth, thumbHeight uint
			if width > height {
				thumbWidth = ThumbSize
				thumbHeight = uint(float64(ThumbSize) / float64(width) * float64(height))
			} else {
				thumbHeight = ThumbSize
				thumbWidth = uint(float64(ThumbSize) / float64(height) * float64(width))
			}
			log.Trace.Printf("Resizing %s thumb image to %dx%d", file.Filename(), thumbWidth, thumbHeight)
			err = mw.ScaleImage(thumbWidth, thumbHeight)
			if err != nil {
				return werror.WithStack(err)
			}
			sw.Lap("Image resize for thumb")
		}

		thumb, err := ms.fileService.NewCacheFile(m, models.LowRes, 0)
		if err != nil && !errors.Is(err, werror.ErrFileAlreadyExists) {
			return werror.WithStack(err)
		} else if err == nil {
			blob, err := mw.GetImageBlob()
			if err != nil {
				return werror.WithStack(err)
			}
			_, err = thumb.Write(blob)
			if err != nil {
				return werror.WithStack(err)
			}
			m.SetLowresCacheFile(thumb)
		}

		sw.Lap("Read raw image")
		sw.Stop()
		// sw.PrintResults(false)
		return nil
	} else if mType.IsVideo() {
		errOut := bytes.NewBuffer(nil)

		const frameNum = 10

		bufbuf := bytes.NewBuffer(buf)
		bufbuf.Reset()

		err = ffmpeg.Input(file.AbsPath()).Filter(
			"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
		).Output(
			"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
		).WithOutput(bufbuf).WithErrorOutput(errOut).Run()
		if err != nil {
			return werror.WithStack(err)
		}
		buf = bufbuf.Bytes()

		sw.Lap("Read video")
	} else {
		fileReader, err := file.Readable()
		if err != nil {
			return err
		}
		bufbuf := bytes.NewBuffer(buf)
		bufbuf.Reset()
		_, err = io.Copy(bufbuf, fileReader)
		if err != nil {
			return err
		}
		buf = bufbuf.Bytes()
		sw.Lap("Read regular image")
	}

	err = ms.generateCacheFiles(m, buf)
	if err != nil {
		return err
	}
	sw.Lap("Generate Cache Files")

	ms.filesBuffer.Put(&buf)

	sw.Lap("Put Buffer")
	sw.Stop()
	// sw.PrintResults(false)

	return nil
}

func (ms *MediaServiceImpl) GetMediaType(m *models.Media) models.MediaType {
	return ms.typeService.ParseMime(m.MimeType)
}

func (ms *MediaServiceImpl) GetMediaTypes() models.MediaTypeService {
	return ms.typeService
}

func (ms *MediaServiceImpl) RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []*models.Media {
	var medias []*models.Media

	for _, f := range folders {
		if f == nil {
			log.Warning.Println("Skipping recursive media lookup for non-existent folder")
			continue
		}
		if !f.IsDir() {
			if ms.IsFileDisplayable(f) {
				m := ms.Get(f.GetContentId())
				if m != nil {
					medias = append(medias, m)
				}
			}
			continue
		}
		err := f.RecursiveMap(
			func(f *fileTree.WeblensFileImpl) error {
				if !f.IsDir() && ms.IsFileDisplayable(f) {
					m := ms.Get(f.GetContentId())
					if m != nil {
						medias = append(medias, m)
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

func (ms *MediaServiceImpl) getFetchMediaCacheImage(ctx context.Context) (data []byte, err error) {
	defer internal.RecoverPanic("Fetching media image had panic")

	m := ctx.Value(CacheMediaKey).(*models.Media)
	q := ctx.Value(CacheQualityKey).(models.MediaQuality)
	pageNum, _ := ctx.Value(CachePageKey).(int)

	f, err := ms.getCacheFile(m, q, pageNum)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, werror.Errorf("This should never happen... file is nil in GetFetchMediaCacheImage")
	}

	data, err = f.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		err = werror.Errorf("displayable bytes empty")
		return nil, err
	}

	return data, nil
}

func (ms *MediaServiceImpl) getCacheFile(
	m *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	if quality == models.LowRes && m.GetLowresCacheFile() != nil {
		return m.GetLowresCacheFile(), nil
	} else if quality == models.HighRes && m.GetHighresCacheFiles(pageNum) != nil {
		return m.GetHighresCacheFiles(pageNum), nil
	}

	filename := m.FmtCacheFileName(quality, pageNum)
	cacheFile, err := ms.fileService.GetMediaCacheByFilename(filename)
	if err != nil {
		return nil, werror.WithStack(werror.ErrNoCache)
	}

	if quality == models.LowRes {
		m.SetLowresCacheFile(cacheFile)
	} else if quality == models.HighRes {
		m.SetHighresCacheFiles(cacheFile, pageNum)
	} else {
		return nil, werror.Errorf("Unknown media quality [%s]", quality)
	}

	return cacheFile, nil
}

func (ms *MediaServiceImpl) generateCacheFiles(m *models.Media, bs []byte) error {
	if len(bs) == 0 {
		return werror.Errorf("empty media buffer")
	}

	var err error
	var rotate int
	if ms.GetMediaType(m).IsRaw() {
		switch m.Rotate {
		case "Rotate 270 CW":
			rotate = 270
		case "Rotate 90 CW":
			rotate = 90
		case "Horizontal (normal)":
		case "":
			log.Debug.Println("empty orientation")
		default:
			err = werror.Errorf("Unknown rotate name [%s]", m.Rotate)
		}
		if err != nil {
			return werror.WithStack(err)
		}
	}

	var imgs = make([][]byte, 0, m.PageCount)
	for i := range m.GetPageCount() {
		opts := bimg.Options{
			Rotate:  bimg.Angle(rotate),
			Type:    bimg.WEBP,
			PageNum: i,
		}

		img, err := bimg.Resize(bs, opts)
		if err != nil {
			return werror.WithStack(err)
		}

		imgs = append(imgs, img)
	}

	imgSize, err := bimg.Size(imgs[0])
	if err != nil {
		return werror.WithStack(err)
	}

	m.Height = imgSize.Height
	m.Width = imgSize.Width

	thumbW := int((models.ThumbnailHeight / float32(m.Height)) * float32(m.Width))

	thumbOpts := bimg.Options{
		Width:  thumbW,
		Height: int(models.ThumbnailHeight),
	}
	thumbBytes, err := bimg.Resize(imgs[0], thumbOpts)
	if err != nil {
		return werror.WithStack(err)
	}

	var thumbFile *fileTree.WeblensFileImpl
	thumbFile, err = ms.fileService.NewCacheFile(m, models.LowRes, 0)
	if err != nil {
		if !errors.Is(err, werror.ErrFileAlreadyExists) {
			return err
		}
	} else {
		_, err = thumbFile.Write(thumbBytes)
		if err != nil {
			return err
		}
	}
	m.SetLowresCacheFile(thumbFile)

	for pageNum := range m.GetPageCount() {
		fullresFile, err := ms.fileService.NewCacheFile(m, models.HighRes, pageNum)
		if err != nil {
			if !errors.Is(err, werror.ErrFileAlreadyExists) {
				return err
			}
		} else {
			_, err = fullresFile.Write(imgs[pageNum])
			if err != nil {
				return err
			}

			m.SetHighresCacheFiles(fullresFile, pageNum)
		}
	}

	return nil
}
