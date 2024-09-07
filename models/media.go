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
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	MediaId primitive.ObjectID `json:"-" bson:"_id"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId ContentId `json:"contentId" bson:"contentId"`

	// Slices of files whos content hash to the contentId
	FileIds []fileTree.FileId `json:"fileIds" bson:"fileIds"`

	CreateDate time.Time `json:"createDate" bson:"createDate"`

	// User who owns the file that resulted in this media being created
	Owner Username `json:"owner" bson:"owner"`

	// Full-res image dimensions
	Width  int `json:"width" bson:"width"`
	Height int `json:"height" bson:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `json:"pageCount" bson:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `json:"duration" bson:"duration"`

	// Unused
	// BlurHash string `json:"blurHash,omitempty" bson:"blurHash,omitempty"`

	// Mime-type key of the media
	MimeType string `json:"mimeType" bson:"mimeType"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `json:"recognitionTags" bson:"recognitionTags"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `json:"hidden" bson:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `json:"enabled" bson:"enabled"`

	LikedBy []Username `json:"likedBy" bson:"likedBy"`

	/* NON-DATABASE FIELDS */

	// Lock to synchronize updates to the media
	updateMu sync.RWMutex

	// The rotation of the image from its original. Found from the exif data
	Rotate string

	// If the media is imported into the databse yet. If not, we shouldn't ask about
	// things like cache, dimensions, etc., as it might not have them.
	imported bool

	// WEBP thumbnail cache fileId
	lowresCacheFile *fileTree.WeblensFileImpl

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	highResCacheFiles []*fileTree.WeblensFileImpl
}

func NewMedia(contentId ContentId) *Media {
	return &Media{
		ContentId: contentId,
	}
}

func (m *Media) ID() ContentId {
	return m.ContentId
}

func (m *Media) SetContentId(id ContentId) {
	m.ContentId = id
}

func (m *Media) IsHidden() bool {
	return m.Hidden
}

func (m *Media) GetCreateDate() time.Time {
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
	return m.FileIds
}

func (m *Media) addFile(f *fileTree.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIds = internal.AddToSet(m.FileIds, f.ID())
}

func (m *Media) removeFile(f *fileTree.WeblensFileImpl) error {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIds = internal.Filter(
		m.FileIds, func(fId fileTree.FileId) bool {
			return fId != f.ID()
		},
	)

	return nil
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

func (m *Media) getImageRecognitionTags() (err error) {
	log.Warning.Println("Skipping image recognition tags")
	return nil
	// bs, err := m.ReadDisplayable(LowRes, 0)
	// if err != nil {
	// 	return
	// }
	// imgBuf := bytes.NewBuffer(bs)
	//
	// resp, err := comm.Post(internal.GetImgRecognitionUrl()+"/recognize", "application/jpeg", imgBuf)
	// if err != nil {
	// 	return
	// }
	// if resp.StatusCode != 200 {
	// 	return fmt.Errorf("failed to get recognition tags: %s", resp.Status)
	// }
	//
	// var recogTags []string
	//
	// err = json.NewDecoder(resp.Body).Decode(&recogTags)
	// if err != nil {
	// 	return err
	// }
	//
	// m.RecognitionTags = recogTags
	//
	// return
}

func (m *Media) UnmarshalBSON(bs []byte) error {
	raw := bson.Raw(bs)
	contentId, ok := raw.Lookup("contentId").StringValueOK()
	if !ok {
		return werror.Errorf("failed to parse contentId")
	}
	m.ContentId = ContentId(contentId)

	filesArr, ok := raw.Lookup("fileIds").ArrayOK()
	if ok {
		fileIdsRaw, err := filesArr.Values()
		if err != nil {
			return werror.WithStack(err)
		}
		m.FileIds = internal.Map(
			fileIdsRaw, func(e bson.RawValue) fileTree.FileId {
				return fileTree.FileId(e.StringValue())
			},
		)
	} else {
		m.FileIds = []fileTree.FileId{}
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

		m.RecognitionTags = internal.Map(
			rts, func(e bson.RawValue) string {
				return e.StringValue()
			},
		)
	}

	hidden, ok := raw.Lookup("hidden").BooleanOK()
	if ok {
		m.Hidden = hidden
	}

	m.imported = true

	return nil
}

func (m *Media) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"contentId":   m.ContentId,
		"fileIds":     m.FileIds,
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

	m.ContentId = ContentId(data["contentId"].(string))

	m.Width = int(data["width"].(float64))
	m.Height = int(data["height"].(float64))
	m.CreateDate = time.UnixMilli(int64(data["createDate"].(float64)))
	m.MimeType = data["mimeType"].(string)

	if data["recognitionTags"] != nil {
		m.RecognitionTags = internal.SliceConvert[string](data["recognitionTags"].([]any))
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

	GetMediaType(m *Media) MediaType
	GetMediaTypes() MediaTypeService
	IsFileDisplayable(file *fileTree.WeblensFileImpl) bool
	IsCached(m *Media) bool
	GetProminentColors(media *Media) (prom []string, err error)

	FetchCacheImg(m *Media, quality MediaQuality, pageNum int) ([]byte, error)
	StreamVideo(m *Media, u *User, share *FileShare) (*VideoStreamer, error)
	StreamCacheVideo(m *Media, startByte, endByte int) ([]byte, error)
	NukeCache() error

	GetFilteredMedia(
		requester *User, sort string, sortDirection int, excludeIds []ContentId, raw bool, hidden bool,
	) ([]*Media, error)
	RecursiveGetMedia(folders ...*fileTree.WeblensFileImpl) []ContentId

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
	media         *Media
	encodingBegun bool
	streamDirPath string
	err           error
}

func NewVideoStreamer(media *Media, destPath string) *VideoStreamer {
	return &VideoStreamer{
		media:         media,
		streamDirPath: destPath,
	}
}

func (vs *VideoStreamer) transcodeChunks(f *fileTree.WeblensFileImpl, speed string) {
	defer func() { vs.encodingBegun = false }()

	err := os.Mkdir(vs.streamDirPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		vs.err = err
		return
	}

	autioRate := 128000
	outErr := bytes.NewBuffer(nil)
	err = ffmpeg.Input(f.GetAbsPath(), ffmpeg.KwArgs{"ss": 0}).Output(
		vs.streamDirPath+"%03d.ts", ffmpeg.KwArgs{
			"c:v":                "libx264",
			"b:v":                400000 * 2,
			"b:a":                autioRate,
			"crf":                18,
			"preset":             speed,
			"segment_list_flags": "+live",
			"format":             "segment",
			"segment_format":     "mpegts",
			"hls_init_time":      5,
			"hls_time":           5,
			"hls_list_size":      0,
			"segment_list":       filepath.Join(vs.streamDirPath, "list.m3u8"),
		},
	).WithErrorOutput(outErr).Run()

	if err != nil {
		log.Error.Println(outErr.String())
		vs.err = err
	}
}

func (vs *VideoStreamer) Encode(f *fileTree.WeblensFileImpl) *VideoStreamer {
	if !vs.encodingBegun {
		vs.encodingBegun = true
		go vs.transcodeChunks(f, "ultrafast")
	}

	return vs
}

func (vs *VideoStreamer) GetEncodeDir() string {
	return vs.streamDirPath
}

func (vs *VideoStreamer) Err() error {
	return vs.err
}

func (vs *VideoStreamer) IsTranscoding() bool {
	return vs.encodingBegun
}

func (vs *VideoStreamer) probeSourceBitrate(f *fileTree.WeblensFileImpl) (int, error) {
	probeJson, err := ffmpeg.Probe(f.GetAbsPath())
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
		return 0, errors.New("invalid movie format")
	}
	bitRate, err := strconv.ParseInt(formatChunk["bit_rate"].(string), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(bitRate), nil
}
