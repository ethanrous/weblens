package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	CreateDate time.Time `bson:"createDate"`

	// WEBP thumbnail cache fileId
	lowresCacheFile *fileTree.WeblensFileImpl

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentID ContentId `bson:"contentId"`

	// User who owns the file that resulted in this media being created
	Owner Username `bson:"owner"`

	// Mime-type key of the media
	MimeType string `bson:"mimeType"`

	// The rotation of the image from its original. Found from the exif data
	Rotate string

	// Slices of files whos content hash to the contentId
	FileIDs []fileTree.FileId `bson:"fileIds"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `bson:"recognitionTags"`

	LikedBy []Username `bson:"likedBy"`

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	highResCacheFiles []*fileTree.WeblensFileImpl

	// Full-res image dimensions
	Width  int `bson:"width"`
	Height int `bson:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `bson:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `bson:"duration"`

	/* NON-DATABASE FIELDS */

	// Lock to synchronize updates to the media
	updateMu sync.RWMutex

	MediaID primitive.ObjectID `bson:"_id" example:"5f9b3b3b7b4f3b0001b3b3b7"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `bson:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `bson:"enabled"`

	// If the media is imported into the databse yet. If not, we shouldn't ask about
	// things like cache, dimensions, etc., as it might not have them.
	imported bool
}

func NewMedia(contentId ContentId) *Media {
	return &Media{
		ContentID: contentId,
	}
}

func (m *Media) ID() ContentId {
	return m.ContentID
}

func (m *Media) SetContentId(id ContentId) {
	m.ContentID = id
}

func (m *Media) IsHidden() bool {
	return m.Hidden
}

func (m *Media) GetCreateDate() time.Time {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.CreateDate
}

func (m *Media) SetCreateDate(t time.Time) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.CreateDate = t
}

func (m *Media) GetPageCount() int {
	return m.PageCount
}

// GetVideoLength returns the length of the media video, if it is a video. Duration is counted in milliseconds
func (m *Media) GetVideoLength() int {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.Duration
}

func (m *Media) SetOwner(owner Username) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.Owner = owner
}

func (m *Media) GetOwner() Username {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.Owner
}

func (m *Media) GetFiles() []fileTree.FileId {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.FileIDs
}

func (m *Media) AddFile(f *fileTree.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIDs = internal.AddToSet(m.FileIDs, f.ID())
}

func (m *Media) RemoveFile(fileIdToRemove fileTree.FileId) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIDs = internal.Filter(
		m.FileIDs, func(fId fileTree.FileId) bool {
			return fId != fileIdToRemove
		},
	)
}

func (m *Media) SetImported(i bool) {
	m.imported = i
}

func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}

	return m.imported
}

func (m *Media) SetEnabled(e bool) {
	m.Enabled = e
}

func (m *Media) IsEnabled() bool {
	return m.Enabled
}

func (m *Media) SetRecognitionTags(tags []string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.RecognitionTags = tags
}

func (m *Media) GetRecognitionTags() []string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.RecognitionTags
}

func (m *Media) setHidden(hidden bool) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.Hidden = hidden
}

func (m *Media) SetLowresCacheFile(thumb *fileTree.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.lowresCacheFile = thumb
}

func (m *Media) GetLowresCacheFile() *fileTree.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.lowresCacheFile
}

func (m *Media) SetHighresCacheFiles(highresFile *fileTree.WeblensFileImpl, pageNum int) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	if len(m.highResCacheFiles) < pageNum+1 {
		m.highResCacheFiles = make([]*fileTree.WeblensFileImpl, m.PageCount)
	}
	m.highResCacheFiles[pageNum] = highresFile
}

func (m *Media) GetHighresCacheFiles(pageNum int) *fileTree.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	if len(m.highResCacheFiles) < pageNum+1 {
		return nil
	}
	return m.highResCacheFiles[pageNum]
}

func (m *Media) FmtCacheFileName(quality MediaQuality, pageNum int) string {
	var pageNumStr string
	if m.PageCount > 1 && quality == HighRes {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}
	filename := fmt.Sprintf("%s-%s%s.cache", m.ID(), quality, pageNumStr)

	return filename
}

const ThumbnailHeight float32 = 500

func (m *Media) UnmarshalBSON(bs []byte) error {
	raw := bson.Raw(bs)
	contentId, ok := raw.Lookup("contentId").StringValueOK()
	if !ok {
		return werror.Errorf("failed to parse contentId")
	}
	m.ContentID = ContentId(contentId)

	filesArr, ok := raw.Lookup("fileIds").ArrayOK()
	if ok {
		fileIdsRaw, err := filesArr.Values()
		if err != nil {
			return werror.WithStack(err)
		}
		m.FileIDs = internal.Map(
			fileIdsRaw, func(e bson.RawValue) fileTree.FileId {
				return fileTree.FileId(e.StringValue())
			},
		)
	} else {
		m.FileIDs = []fileTree.FileId{}
	}

	m.PageCount = int(raw.Lookup("pageCount").Int32())

	videoLength, ok := raw.Lookup("videoLength").Int32OK()
	if ok {
		m.Duration = int(videoLength)
	}

	// m.BlurHash = raw.Lookup("blurHash").StringValue()
	m.Owner = Username(raw.Lookup("owner").StringValue())
	m.Width = int(raw.Lookup("width").Int32())
	m.Height = int(raw.Lookup("height").Int32())

	duration, ok := raw.Lookup("duration").Int32OK()
	if ok {
		m.Duration = int(duration)
	}

	create := raw.Lookup("createDate")
	createTime, ok := create.TimeOK()
	if !ok {
		m.CreateDate = time.UnixMilli(create.Int64())
	} else {
		m.CreateDate = createTime
	}
	m.MimeType = raw.Lookup("mimeType").StringValue()

	likedArr, ok := raw.Lookup("likedBy").ArrayOK()
	if ok {
		likedValues, err := likedArr.Values()
		if err != nil {
			return werror.WithStack(err)
		}
		m.LikedBy = internal.Map(
			likedValues, func(e bson.RawValue) Username {
				return Username(e.StringValue())
			},
		)
	} else {
		m.LikedBy = []Username{}
	}

	rtArr, ok := raw.Lookup("recognitionTags").ArrayOK()
	if ok {
		rts, err := rtArr.Values()
		if err != nil {
			return werror.WithStack(err)
		}

		m.SetRecognitionTags(internal.Map(
			rts, func(e bson.RawValue) string {
				return e.StringValue()
			},
		))
	}

	hidden, ok := raw.Lookup("hidden").BooleanOK()
	if ok {
		m.setHidden(hidden)
	}

	m.imported = true

	return nil
}

func (m *Media) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"contentId":   m.ContentID,
		"fileIds":     m.FileIDs,
		"owner":       m.Owner,
		"width":       m.Width,
		"height":      m.Height,
		"createDate":  m.CreateDate.UnixMilli(),
		"mimeType":    m.MimeType,
		"pageCount":   m.PageCount,
		"imported":    m.imported,
		"hidden":      m.Hidden,
		"likedBy":     m.LikedBy,
		"videoLength": m.Duration,
	}

	return json.Marshal(data)
}

func (m *Media) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return werror.WithStack(err)
	}

	m.ContentID = ContentId(data["contentId"].(string))

	m.Width = int(data["width"].(float64))
	m.Height = int(data["height"].(float64))
	m.CreateDate = time.UnixMilli(int64(data["createDate"].(float64)))
	m.MimeType = data["mimeType"].(string)

	if data["recognitionTags"] != nil {
		m.SetRecognitionTags(internal.SliceConvert[string](data["recognitionTags"].([]any)))
	}

	m.PageCount = int(data["pageCount"].(float64))
	m.imported = data["imported"].(bool)
	m.Hidden = data["hidden"].(bool)

	if data["videoLength"] != nil {
		m.Duration = int(data["videoLength"].(float64))
	}

	return nil
}

type MediaService interface {
	Size() int
	Add(media *Media) error

	Get(id ContentId) *Media
	GetAll() []*Media

	Del(id ContentId) error
	HideMedia(m *Media, hidden bool) error
	AdjustMediaDates(anchor *Media, newTime time.Time, extraMedias []*Media) error

	LoadMediaFromFile(m *Media, file *fileTree.WeblensFileImpl) error
	RemoveFileFromMedia(media *Media, fileId fileTree.FileId) error
	Cleanup() error
	Drop() error
	AddFileToMedia(media *Media, file *fileTree.WeblensFileImpl) error

	GetMediaType(m *Media) MediaType
	GetMediaTypes() MediaTypeService
	IsFileDisplayable(file *fileTree.WeblensFileImpl) bool
	IsCached(m *Media) bool
	GetProminentColors(media *Media) (prom []string, err error)
	GetMediaConverted(m *Media, format string) ([]byte, error)

	FetchCacheImg(m *Media, quality MediaQuality, pageNum int) ([]byte, error)
	StreamVideo(m *Media, u *User, share *FileShare) (*VideoStreamer, error)
	StreamCacheVideo(m *Media, startByte, endByte int) ([]byte, error)

	GetFilteredMedia(
		requester *User, sort string, sortDirection int, excludeIds []ContentId, raw bool, hidden bool, search string,
	) ([]*Media, error)
	RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []*Media

	SetMediaLiked(mediaId ContentId, liked bool, username Username) error
}

type ContentId = string
type MediaQuality string

const (
	LowRes  MediaQuality = "thumbnail"
	HighRes MediaQuality = "fullres"
	Video   MediaQuality = "video"
)

type VideoStreamer struct {
	err           error
	file          *fileTree.WeblensFileImpl
	streamDirPath string
	listFileCache []byte
	updateMu      sync.RWMutex
	encodingBegun atomic.Bool

	log *zerolog.Logger
}

func NewVideoStreamer(file *fileTree.WeblensFileImpl, thumbsPath string) *VideoStreamer {
	destPath := fmt.Sprintf("%s/%s-stream/", thumbsPath, file.GetContentId())

	return &VideoStreamer{
		file:          file,
		streamDirPath: destPath,
	}
}

func (vs *VideoStreamer) transcodeChunks(f *fileTree.WeblensFileImpl, speed string) {
	defer func() {
		vs.encodingBegun.Store(false)
		e := recover()
		if e == nil {
			return
		}

		err, ok := e.(error)
		if !ok {
			vs.log.Error().Msgf("transcodeChunks panicked and got non-error error: %v", e)
			return
		}
		vs.log.Error().Stack().Err(err).Msg("")
	}()

	vs.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Transcoding video %s => %s", f.AbsPath(), vs.streamDirPath) })

	err := os.Mkdir(vs.streamDirPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()
		return
	}

	videoBitrate, audioBitrate, err := vs.probeSourceBitrate(f)
	if err != nil {
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()
		return
	}

	vs.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Bitrate: %d %d", videoBitrate, audioBitrate) })
	outputArgs := ffmpeg.KwArgs{
		"c:v":                "libx264",
		"b:v":                int(videoBitrate),
		"b:a":                320_000,
		"c:a":                "aac",
		"segment_list_flags": "+live",
		"format":             "segment",
		"segment_format":     "mpegts",
		"hls_init_time":      5,
		"hls_time":           5,
		"hls_list_size":      0,
		"segment_list":       filepath.Join(vs.streamDirPath, "list.m3u8"),
		"crf":                18,
		"preset":             speed,
	}

	outErr := bytes.NewBuffer(nil)
	err = ffmpeg.Input(f.AbsPath(), ffmpeg.KwArgs{"ss": 0}).Output(vs.streamDirPath+"%03d.ts", outputArgs).WithErrorOutput(outErr).Run()

	if err != nil {
		log.Error.Println(outErr.String())
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()
	}
}

func (vs *VideoStreamer) Encode(f *fileTree.WeblensFileImpl) *VideoStreamer {
	if !vs.encodingBegun.Load() {
		vs.encodingBegun.Store(true)
		go vs.transcodeChunks(f, "ultrafast")
	}

	return vs
}

func (vs *VideoStreamer) GetEncodeDir() string {
	return vs.streamDirPath
}

func (vs *VideoStreamer) GetChunk(chunkName string) (*os.File, error) {
	chunkPath := filepath.Join(vs.GetEncodeDir(), chunkName)
	if _, err := os.Stat(chunkPath); err != nil {
		vs.Encode(vs.file)

		for vs.IsTranscoding() {
			if _, err := os.Stat(chunkPath); err == nil {
				break
			}
			if vs.Err() != nil {
				return nil, vs.Err()
			}
			time.Sleep(time.Second)
		}
	}

	return os.Open(chunkPath)
}

func (vs *VideoStreamer) GetListFile() ([]byte, error) {
	if vs.listFileCache != nil {
		return vs.listFileCache, nil
	}

	listPath := filepath.Join(vs.GetEncodeDir(), "list.m3u8")
	if _, err := os.Stat(listPath); err != nil {
		vs.Encode(vs.file)

		for vs.IsTranscoding() {
			if _, err := os.Stat(listPath); err == nil {
				break
			}
			if vs.Err() != nil {
				return nil, vs.Err()
			}
			time.Sleep(time.Second)
		}
	}

	listFile, err := os.ReadFile(listPath)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	// Cache the list file only if transcoding is done
	// if bytes.HasSuffix(listFile, []byte("ENDLIST")) {
	// 	vs.listFileCache = listFile
	// }

	// Cache the list file only if transcoding is finished and no errors
	if !vs.IsTranscoding() && vs.Err() == nil {
		vs.listFileCache = listFile
	}

	return listFile, nil
}

func (vs *VideoStreamer) Err() error {
	vs.updateMu.RLock()
	defer vs.updateMu.RUnlock()
	return vs.err
}

func (vs *VideoStreamer) IsTranscoding() bool {
	return vs.encodingBegun.Load()
}

func (vs *VideoStreamer) probeSourceBitrate(f *fileTree.WeblensFileImpl) (videoBitrate int64, audioBitrate int64, err error) {
	vs.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Probing %s", f.AbsPath()) })
	probeJson, err := ffmpeg.Probe(f.AbsPath())
	if err != nil {
		return 0, 0, err
	}
	probeResult := map[string]any{}
	err = json.Unmarshal([]byte(probeJson), &probeResult)
	if err != nil {
		return 0, 0, err
	}

	formatChunk, ok := probeResult["format"].(map[string]any)
	if !ok {
		return 0, 0, werror.Errorf("invalid movie format")
	}

	streamsChunk, ok := probeResult["streams"].([]any)
	if !ok {
		return 0, 0, werror.Errorf("invalid movie format")
	}

	bitRateStr, ok := formatChunk["bit_rate"].(string)
	if !ok {
		return 0, 0, werror.Errorf("bitrate does not exist or is not a string")
	}
	videoBitrate, err = strconv.ParseInt(bitRateStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	audioBitrate = 320_000
	for _, stream := range streamsChunk {
		streamMap := stream.(map[string]any)
		if streamMap["codec_type"].(string) == "audio" {
			bitRate, ok := streamMap["bit_rate"].(string)
			if !ok {
				continue
			}
			audioBitrate, err = strconv.ParseInt(bitRate, 10, 64)
			if err != nil {
				return 0, 0, err
			}
			break
		}
	}

	return videoBitrate, audioBitrate, nil
}
