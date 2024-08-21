package weblens

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/pkg/errors"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	MediaId primitive.ObjectID `json:"-" bson:"_id"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId ContentId `json:"mediaId" bson:"mediaId"`

	// Slices of files whos content hash to the contentId
	FileIds []fileTree.FileId `json:"fileIds" bson:"fileIds"`

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	FullresCacheIds []fileTree.FileId `json:"fullresCacheIds" bson:"fullresCacheIds"`

	// WEBP thumbnail cache fileId
	ThumbnailCacheId fileTree.FileId `json:"thumbnailCacheId" bson:"thumbnailCacheId"`

	CreateDate time.Time `json:"createDate" bson:"createDate"`

	// User who owns the file that resulted in this media being created
	Owner *User `json:"owner" bson:"owner"`

	// Full-res image dimentions
	MediaWidth  int `json:"mediaWidth" bson:"mediaWidth"`
	MediaHeight int `json:"mediaHeight" bson:"mediaHeight"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `json:"pageCount" bson:"pageCount"`

	// Total time, in milliseconds, of a video
	VideoLength int `json:"videoLength" bson:"videoLength"`

	// Unused
	BlurHash string `json:"blurHash" bson:"blurHash"`

	// Mime-type key of the media
	MimeType string `json:"mimeType" bson:"mimeType"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `json:"recognitionTags" bson:"recognitionTags"`

	// If the media is hidden from the timeline
	Hidden bool `json:"hidden" bson:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `json:"enabled" bson:"enabled"`

	LikedBy []Username `json:"likedBy" bson:"likedBy"`

	/* NON-DATABASE FIELDS */

	// Lock to synchronize updates to the media
	updateMu sync.RWMutex

	// Real media type of the media, loaded from the MimeType
	mediaType MediaType

	// If the media is imported into the databse yet. If not, we shouldn't ask about
	// things like cache, dimentions, etc., as it might not have them.
	imported bool

	// Result of exif scan of the real file. Cleared after it is read
	rawExif map[string]any

	// The rotation of the image from its original. Found from the exif data
	rotate string

	// Bytes of the image(s) being processed into WEBP format for caching
	image  *bimg.Image
	images []*bimg.Image

	// Live file of the cache that supports this media.
	thumbCacheFile    *fileTree.WeblensFile
	fullresCacheFiles []*fileTree.WeblensFile

	streamer *media.VideoStreamer
}

func New(contentId ContentId) *Media {
	return &Media{
		ContentId: contentId,
	}
}

func (m *Media) LoadFromFile(f *fileTree.WeblensFile, preReadBytes []byte, task types.Task) (
	retM *Media, err error,
) {
	err = m.parseExif(f)
	if err != nil {
		return
	}
	task.SwLap("Parse Exif")

	err = m.AddFile(f)
	if err != nil {
		return nil, err
	}

	err = m.handleCacheCreation(f)
	if err != nil {
		return
	}
	task.SwLap("Create cache")

	if m.RecognitionTags == nil && m.mediaType.SupportsImgRecog() {
		err = m.getImageRecognitionTags()
		if err != nil {
			wlog.ErrTrace(err)
		}
		task.SwLap("Get img recognition tags")
	}

	m.SetOwner(f.Owner())
	task.SwLap("Add file and set owner")

	// if m.MediaType.IsMultiPage() {
	// 	err = m.generateImages()
	// } else {
	// 	err = m.generateImage()
	// }
	// if err != nil {
	// 	util.ErrTrace(err, f.GetAbsPath())
	// 	return
	// }
	// task.SwLap("Generate Image")

	return m, nil
}

func (m *Media) ID() ContentId {
	return m.ContentId
}

func (m *Media) SetContentId(id ContentId) {
	m.ContentId = id
}

func (m *Media) IsFilledOut() (bool, string) {
	if m.ContentId == "" {
		return false, "mediaId"
	}
	if len(m.FileIds) == 0 {
		return false, "file id"
	}
	if m.Owner == nil {
		return false, "owner"
	}
	if m.mediaType.SupportsImgRecog() && m.RecognitionTags == nil {
		return false, "recognition tags"
	}

	// Visual Media specific properties
	if m.mediaType != nil && m.mediaType.IsDisplayable() {
		// if m.BlurHash == "" {
		// 	return false, "blurhash"
		// }
		if m.MediaWidth == 0 {
			return false, "Media width"
		}
		if m.MediaHeight == 0 {
			return false, "Media height"
		}
		// if m.ThumbWidth == 0 {
		// 	return false, "thumb width"
		// }
		// if m.ThumbHeight == 0 {
		// 	return false, "thumb height"
		// }

	}

	if m.CreateDate.IsZero() {
		return false, "create date"
	}

	return true, ""
}

func (m *Media) IsHidden() bool {
	return m.Hidden
}

func (m *Media) GetCreateDate() time.Time {
	return m.CreateDate
}

func (m *Media) SetCreateDate(t time.Time) error {
	// err := dataStore.dbServer.adjustMediaDate(m, t)
	// if err != nil {
	// 	return err
	// }
	// m.CreateDate = t

	// return nil

	return werror.NotImplemented("Media Set Create Date")
}

func (m *Media) GetMediaType() *MediaType {
	if m.mediaType != nil {
		wlog.Debug.Println("MEDIA TYPE", m.mediaType)
	}
	if m.mediaType != nil && m.mediaType.GetMime() != "" {
		return m.mediaType
	}
	if m.MimeType != "" {
		m.mediaType = types.SERV.MediaRepo.TypeService().ParseMime(m.MimeType)
	}
	return m.mediaType
}

func (m *Media) GetPageCount() int {
	return m.PageCount
}

// GetVideoLength returns the length of the media video, if it is a video. Duration is counted in milliseconds
func (m *Media) GetVideoLength() int {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.VideoLength
}

func (m *Media) SetOwner(owner *User) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.Owner = owner
}

func (m *Media) GetOwner() *User {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	if m.Owner == nil {
		wlog.Error.Println("Media owner is nil")
		return nil
	}
	return m.Owner
}

func (m *Media) ReadDisplayable(q MediaQuality, index int) (data []byte, err error) {
	return types.SERV.MediaRepo.FetchCacheImg(m, q, index)
}

func (m *Media) GetFiles() []fileTree.FileId {
	return m.FileIds
}

func (m *Media) getExistingFiles() []fileTree.FileId {
	return internal.Filter(
		m.FileIds, func(fId fileTree.FileId) bool {
			return types.SERV.FileTree.Get(fId) != nil
		},
	)
}

func (m *Media) AddFile(f *fileTree.WeblensFile) error {
	m.FileIds = internal.AddToSet(m.FileIds, f.ID())
	if m.IsImported() {
		return types.SERV.StoreService.AddFileToMedia(m.ID(), f.ID())
	}
	return nil
}

func (m *Media) RemoveFile(f *fileTree.WeblensFile) error {
	var existed bool
	m.FileIds, _, existed = internal.YoinkFunc(
		m.FileIds, func(existFile fileTree.FileId) bool { return existFile == f.ID() },
	)

	if !existed {
		wlog.Warning.Printf("Attempted to remove file %s from Media that did not have that file", f)
	}

	err := types.SERV.StoreService.RemoveFileFromMedia(m.ID(), f.ID())
	if err != nil {
		return err
	}

	if len(m.FileIds) == 0 {
		err := m.Hide(true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Clean Toss the cached data for the Media generated when parsing a file.
// This will drastically reduce memory usage if used properly
func (m *Media) Clean() {
	if m == nil {
		return
	}
	// return

	m.rawExif = nil
	m.image = nil
	m.images = nil
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

func (m *Media) Hide(hidden bool) error {
	m.Hidden = hidden
	return types.SERV.StoreService.SetMediaHidden(m.ID(), hidden)
}

func (m *Media) GetProminentColors() (prom []string, err error) {
	var i image.Image
	thumbBytes, err := m.ReadDisplayable(Thumbnail, 0)
	if err != nil {
		return
	}

	i, err = webp.Decode(bytes.NewBuffer(thumbBytes), nil)
	if err != nil {
		return
	}

	promColors, err := prominentcolor.Kmeans(i)
	prom = internal.Map(promColors, func(p prominentcolor.ColorItem) string { return p.AsString() })
	return
}

// Private

func (m *Media) loadExif(f *fileTree.WeblensFile) error {
	fileInfos, err := types.SERV.MediaRepo.RunExif(f.GetAbsPath())
	if err != nil {
		return err
	}
	if fileInfos[0].Err != nil {
		return fileInfos[0].Err
	}

	m.rawExif = fileInfos[0].Fields
	return nil
}

var timeOrigin = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

func (m *Media) parseExif(f *fileTree.WeblensFile, mediaTypeService MediaTypeService) error {

	if m.CreateDate.Unix() == 0 && m.mediaType != nil {
		return nil
	}

	if m.rawExif == nil {
		err := m.loadExif(f)
		if err != nil {
			return err
		}
	}

	var err error
	if m.CreateDate.Unix() <= 0 {
		r, ok := m.rawExif["SubSecCreateDate"]
		if !ok {
			r, ok = m.rawExif["MediaCreateDate"]
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
				m.CreateDate = f.ModTime()
			}
		} else {
			m.CreateDate = f.ModTime()
		}
	}

	if m.mediaType == nil {
		mimeType, ok := m.rawExif["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MimeType = mimeType
		m.mediaType = mediaTypeService.ParseMime(mimeType)

		if m.mediaType == nil {
			return errors.Errorf("failed to parse media type: %s", mimeType)
		}

		if m.mediaType.IsVideo() {
			probeJson, err := ffmpeg.Probe(f.GetAbsPath())
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

	if m.mediaType.IsMultiPage() {
		m.PageCount = int(m.rawExif["PageCount"].(float64))
	} else {
		m.PageCount = 1
	}

	if len(m.FullresCacheIds) != m.PageCount {
		m.FullresCacheIds = make([]fileTree.FileId, m.PageCount)
	}

	if len(m.fullresCacheFiles) != m.PageCount {
		m.fullresCacheFiles = make([]*fileTree.WeblensFile, m.PageCount)
	}

	if m.rotate == "" {
		rotate := m.rawExif["Orientation"]
		if rotate != nil {
			m.rotate = rotate.(string)
		}
	}

	if !m.mediaType.IsDisplayable() || m.mediaType.GetThumbExifKey() == "" {
		return nil
	}

	return nil
}

func (m *Media) generateImage(bs []byte) (err error) {
	m.image = bimg.NewImage(bs)

	if m.rotate == "" {
		err = m.parseExif(types.SERV.FileTree.Get(m.getExistingFiles()[0]))
		if err != nil {
			return err
		}
	}

	// Rotation is backwards because bimg rotates CW, but exif stores CW rotation
	if m.GetMediaType().IsRaw() {
		switch m.rotate {
		case "Rotate 270 CW":
			_, err = m.image.Rotate(270)
		case "Rotate 90 CW":
			_, err = m.image.Rotate(90)
		case "Horizontal (normal)":
		case "":
			wlog.Debug.Println("empty orientation")
		default:
			err = error2.NewWeblensError(fmt.Sprintf("Unknown rotate name [%s]", m.rotate))
		}
		if err != nil {
			return
		}
	}

	_, err = m.image.Convert(bimg.WEBP)
	if err != nil {
		return
	}

	imgSize, err := m.image.Size()
	if err != nil {
		return
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	return nil
}

func (m *Media) generateImages(bs []byte) (err error) {
	if !m.mediaType.IsMultiPage() {
		return error2.NewWeblensError("cannot load multi-page image without page count")
	}

	m.images = make([]*bimg.Image, m.PageCount)
	var bi *bimg.Image
	for page := range m.PageCount {
		bi = bimg.NewImage(bs)
		pageBytes, err := bi.Process(bimg.Options{Type: bimg.WEBP, PageNum: page})
		if err != nil {
			return err
		}
		m.images[page] = bimg.NewImage(pageBytes)
	}

	imgSize, err := bi.Size()
	if err != nil {
		return
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	return nil
}

func (m *Media) GetCacheFile(q MediaQuality, generateIfMissing bool, pageNum int) (f *fileTree.WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	} else if q == Fullres && len(m.fullresCacheFiles) > pageNum && m.fullresCacheFiles[pageNum] != nil {
		f = m.fullresCacheFiles[pageNum]
		return
	}

	if pageNum >= m.PageCount {
		return nil, dataStore.ErrPageOutOfRange
	}

	cacheRoot := types.SERV.FileTree.Get("CACHE")

	var cacheFileId fileTree.FileId
	if q == Fullres && len(m.FullresCacheIds) > pageNum && m.FullresCacheIds[pageNum] != "" {
		cacheFileId = m.FullresCacheIds[pageNum]
	} else if q == Fullres {
		cacheFileId = m.getCacheId(q, pageNum, cacheRoot)
		m.FullresCacheIds = make([]fileTree.FileId, m.PageCount)
		m.FullresCacheIds[pageNum] = cacheFileId
	}

	if q == Thumbnail {
		if m.ThumbnailCacheId == "" {
			m.ThumbnailCacheId = m.getCacheId(q, pageNum, cacheRoot)
		}
		cacheFileId = m.ThumbnailCacheId
	}

	f = types.SERV.FileTree.Get(cacheFileId)
	if f == nil || !f.Exists() {
		if generateIfMissing {
			realFile := types.SERV.FileTree.Get(m.getExistingFiles()[0])
			if realFile == nil {
				return nil, dataStore.ErrNoCache()
			}

			wlog.Warning.Printf("Cache file F[%s] not found for M[%s], generating...", cacheFileId, m.ID())
			err = m.handleCacheCreation(realFile)
			if err != nil {
				return
			}

			return m.GetCacheFile(q, false, pageNum)

		} else {
			return nil, dataStore.ErrNoCache()
		}
	}

	if q == Thumbnail {
		m.thumbCacheFile = f
	} else if q == Fullres {
		m.fullresCacheFiles[pageNum] = f
	}

	return
}

const ThumbnailHeight float32 = 500

func (m *Media) handleCacheCreation(f *fileTree.WeblensFile) (err error) {
	_, err = m.GetCacheFile(Thumbnail, false, 0)
	hasThumbCache := err == nil
	_, err = m.GetCacheFile(Fullres, false, 0)
	hasFullresCache := err == nil

	if hasThumbCache && hasFullresCache && m.MediaWidth != 0 && m.MediaHeight != 0 {
		return nil
	}

	var bs []byte

	if m.mediaType.IsRaw() {
		if m.rawExif == nil {
			err = m.loadExif(f)
			if err != nil {
				return err
			}
		}
		raw64 := m.rawExif[m.mediaType.GetThumbExifKey()].(string)
		raw64 = raw64[strings.Index(raw64, ":")+1:]

		imgBytes, err := base64.StdEncoding.DecodeString(raw64)
		if err != nil {
			return err
		}
		bs = imgBytes
	} else if m.mediaType.IsVideo() {
		out := bytes.NewBuffer(nil)
		errOut := bytes.NewBuffer(nil)

		const frameNum = 10

		err = ffmpeg.Input(f.GetAbsPath()).Filter(
			"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
		).Output(
			"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
		).WithOutput(out).WithErrorOutput(errOut).Run()
		if err != nil {
			wlog.Error.Println(errOut.String())
			return err
		}
		bs = out.Bytes()

	} else {
		bs, err = f.ReadAll()
		if err != nil {
			return err
		}
	}

	if m.image == nil && !m.mediaType.IsMultiPage() {
		err = m.generateImage(bs)
		if err != nil {
			return
		}
	} else if m.mediaType.IsMultiPage() && (len(m.images) == 0 || m.images[0] == nil) {
		err = m.generateImages(bs)
		if err != nil {
			return
		}
	}

	if !m.mediaType.IsMultiPage() && m.image == nil || m.mediaType.IsMultiPage() && m.images == nil {
		return dataStore.ErrNoImage
	}

	thumbW := int((ThumbnailHeight / float32(m.MediaHeight)) * float32(m.MediaWidth))

	var thumbBytes []byte
	if !m.mediaType.IsMultiPage() && m.image != nil {

		// Copy image buffer for the thumbnail
		thumbImg := bimg.NewImage(m.image.Image())

		thumbBytes, err = thumbImg.Resize(thumbW, int(ThumbnailHeight))
		if err != nil {
			return
		}
		thumbSize, err := thumbImg.Size()
		if err != nil {
			wlog.ShowErr(err)
		} else {
			thumbRatio := float64(thumbSize.Width) / float64(thumbSize.Height)
			mediaRatio := float64(m.MediaWidth) / float64(m.MediaHeight)
			// util.Debug.Println(thumbRatio, mediaRatio)
			if (thumbRatio < 1 && mediaRatio > 1) || (thumbRatio > 1 && mediaRatio < 1) {
				wlog.Error.Println("Mismatched media sizes")
			}
		}
	}

	if m.mediaType.IsMultiPage() && len(m.images) != 0 && m.images[0] != nil {
		thumbImg := bimg.NewImage(m.images[0].Image())

		thumbBytes, err = thumbImg.Resize(thumbW, int(ThumbnailHeight))
		if err != nil {
			return
		}
	}

	if !hasThumbCache {
		m.cacheDisplayable(Thumbnail, thumbBytes, 0, f.GetTree())
	}

	if !hasFullresCache {
		if m.mediaType.IsMultiPage() {
			for page := range m.PageCount {
				m.cacheDisplayable(Fullres, m.images[page].Image(), page, f.GetTree())
			}
		} else {
			m.cacheDisplayable(Fullres, m.image.Image(), 0, f.GetTree())
		}
	}

	return
}

func (m *Media) cacheDisplayable(q MediaQuality, data []byte, pageNum int, ft types.FileTree) *fileTree.WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

	if len(data) == 0 {
		wlog.ErrTrace(errors.New("no data while trying to cache displayable"))
		return nil
	}

	cacheRoot := ft.Get("CACHE")
	f, err := ft.Touch(cacheRoot, cacheFileName, false, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "file already exists") {
			wlog.ErrTrace(err)
			return nil
		} else {
			return f
		}
	}

	err = f.Write(data)
	if err != nil {
		wlog.ErrTrace(err)
		return f
	}

	if len(m.FullresCacheIds) != m.PageCount {
		m.FullresCacheIds = make([]fileTree.FileId, m.PageCount)
	}

	if q == Thumbnail && m.ThumbnailCacheId == "" || m.thumbCacheFile == nil {
		m.ThumbnailCacheId = f.ID()
		m.thumbCacheFile = f
	} else if q == Fullres && m.FullresCacheIds[pageNum] == "" || m.fullresCacheFiles[pageNum] == nil {
		m.FullresCacheIds[pageNum] = f.ID()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *Media) getCacheId(q MediaQuality, pageNum int, cacheDir *fileTree.WeblensFile) fileTree.FileId {
	return cacheDir.GetTree().GenerateFileId(
		filepath.Join(
			cacheDir.GetAbsPath(), m.getCacheFilename(
				q,
				pageNum,
			),
		),
	)
}

func (m *Media) getCacheFilename(q MediaQuality, pageNum int) string {
	var cacheFileName string

	if m.PageCount == 1 || q == Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.cache", m.ID(), q)
	} else if q != Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.cache", m.ID(), q, pageNum)
	}

	return cacheFileName
}

func (m *Media) getImageRecognitionTags() (err error) {
	wlog.Warning.Println("Skipping image recognition tags")
	return nil
	bs, err := m.ReadDisplayable(Thumbnail, 0)
	if err != nil {
		return
	}
	imgBuf := bytes.NewBuffer(bs)

	resp, err := http.Post(internal.GetImgRecognitionUrl()+"/recognize", "application/jpeg", imgBuf)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get recognition tags: %s", resp.Status)
	}

	var recogTags []string

	err = json.NewDecoder(resp.Body).Decode(&recogTags)
	if err != nil {
		return err
	}

	m.RecognitionTags = recogTags

	return
}

func (m *Media) MarshalBSON() ([]byte, error) {
	data := map[string]any{
		"contentId":        m.ContentId,
		"fileIds":          m.FileIds,
		"thumbnailCacheId": m.ThumbnailCacheId,
		"fullresCacheIds":  m.FullresCacheIds,
		"blurHash":         m.BlurHash,
		"owner":            m.Owner.GetUsername(),
		"width":            m.MediaWidth,
		"height":           m.MediaHeight,
		"createDate":       m.CreateDate.UnixMilli(),
		"mimeType":         m.MimeType,
		"recognitionTags":  m.RecognitionTags,
		"pageCount":        m.PageCount,
		"videoLength":      m.VideoLength,
	}

	return bson.Marshal(data)
}

func (m *Media) UnmarshalBSON(bs []byte) error {
	raw := bson.Raw(bs)
	contentId, ok := raw.Lookup("contentId").StringValueOK()
	if !ok {
		return error2.WErrMsg("failed to parse contentId")
	}
	m.ContentId = ContentId(contentId)

	fileIds, err := raw.Lookup("fileIds").Array().Elements()
	if err != nil {
		return error2.Wrap(err)
	}
	m.FileIds = internal.Map(
		fileIds, func(e bson.RawElement) fileTree.FileId {
			return fileTree.FileId(e.Value().StringValue())
		},
	)

	m.ThumbnailCacheId = fileTree.FileId(raw.Lookup("thumbnailCacheId").StringValue())

	m.PageCount = int(raw.Lookup("pageCount").Int32())

	videoLength, ok := raw.Lookup("videoLength").Int32OK()
	if ok {
		m.VideoLength = int(videoLength)
	}

	fullresIds, err := raw.Lookup("fullresCacheIds").Array().Elements()
	if err != nil {
		return error2.Wrap(err)
	}
	m.FullresCacheIds = internal.Map(
		fullresIds, func(e bson.RawElement) fileTree.FileId {
			return fileTree.FileId(e.Value().StringValue())
		},
	)

	m.fullresCacheFiles = internal.Map(
		m.FullresCacheIds, func(fId fileTree.FileId) *fileTree.WeblensFile {
			return types.SERV.FileTree.Get(fId)
		},
	)

	// TODO - figure out why this happens
	if len(m.fullresCacheFiles) != m.PageCount {
		wlog.Warning.Printf("Bad fullres file count, got %d but expected %d", len(m.fullresCacheFiles), m.PageCount)
		m.fullresCacheFiles = make([]*fileTree.WeblensFile, m.PageCount)
	}

	m.BlurHash = raw.Lookup("blurHash").StringValue()
	// m.Owner = types.SERV.UserService.Get(types.Username(raw.Lookup("owner").StringValue()))
	m.MediaWidth = int(raw.Lookup("width").Int32())
	m.MediaHeight = int(raw.Lookup("height").Int32())
	m.CreateDate = time.UnixMilli(raw.Lookup("createDate").Int64())
	m.MimeType = raw.Lookup("mimeType").StringValue()

	likedArr, ok := raw.Lookup("likedBy").ArrayOK()
	if ok {
		likedValues, err := likedArr.Values()
		if err != nil {
			return error2.Wrap(err)
		}
		m.LikedBy = internal.Map(
			likedValues, func(e bson.RawValue) types.Username {
				return types.Username(e.StringValue())
			},
		)
	} else {
		m.LikedBy = []types.Username{}
	}

	rtArr, ok := raw.Lookup("recognitionTags").ArrayOK()
	if ok {
		rts, err := rtArr.Values()
		if err != nil {
			return error2.Wrap(err)
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
		"contentId":        m.ContentId,
		"fileIds":          m.FileIds,
		"thumbnailCacheId": m.ThumbnailCacheId,
		"fullresCacheIds":  m.FullresCacheIds,
		"owner":       m.Owner.GetUsername(),
		"width":       m.MediaWidth,
		"height":      m.MediaHeight,
		"createDate":  m.CreateDate.UnixMilli(),
		"mimeType":    m.MimeType,
		"pageCount":   m.PageCount,
		"imported":    m.imported,
		"hidden":      m.Hidden,
		"likedBy":     m.LikedBy,
		"videoLength": m.VideoLength,
		// "blurHash":         m.BlurHash,
		// "recognitionTags":  m.RecognitionTags,
	}

	return json.Marshal(data)
}

func (m *Media) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return errors.WithStack(err)
	}

	m.ContentId = ContentId(data["contentId"].(string))
	m.ThumbnailCacheId = fileTree.FileId(data["thumbnailCacheId"].(string))

	idsStrings := internal.SliceConvert[string](data["fullresCacheIds"].([]any))

	m.FullresCacheIds = internal.Map(
		idsStrings, func(s string) fileTree.FileId {
			return fileTree.FileId(s)
		},
	)
	if data["blurHash"] != nil {
		m.BlurHash = data["blurHash"].(string)
	}
	// m.Owner = types.SERV.UserService.Get(types.Username(data["owner"].(string)))
	m.MediaWidth = int(data["width"].(float64))
	m.MediaHeight = int(data["height"].(float64))
	m.CreateDate = time.UnixMilli(int64(data["createDate"].(float64)))
	m.MimeType = data["mimeType"].(string)

	if data["recognitionTags"] != nil {
		m.RecognitionTags = internal.SliceConvert[string](data["recognitionTags"].([]any))
	}

	m.PageCount = int(data["pageCount"].(float64))
	m.imported = data["imported"].(bool)
	m.Hidden = data["hidden"].(bool)

	if data["videoLength"] != nil {
		m.VideoLength = int(data["videoLength"].(float64))
	}

	return nil
}

type MediaService interface {
	Init(store types.AlbumsStore) error
	Size() int
	Get(id ContentId) *Media
	Add(media *Media) error
	Del(id ContentId) error

	TypeService() MediaTypeService
	FetchCacheImg(m *Media, quality MediaQuality, pageNum int) ([]byte, error)
	StreamCacheVideo(m *Media, startByte, endByte int) ([]byte, error)
	GetFilteredMedia(
		requester User, sort string, sortDirection int, albumFilter []AlbumId, raw bool, hidden bool,
	) ([]*Media, error)
	RunExif(path string) ([]exiftool.FileMetadata, error)
	GetAll() []*Media
	NukeCache() error
	SetMediaLiked(mediaId ContentId, liked bool, username Username) error
}

type ContentId string
type MediaQuality string

const (
	Thumbnail MediaQuality = "thumbnail"
	Fullres   MediaQuality = "fullres"
	Video     MediaQuality = "video"
)
