package media

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethanrous/bimg"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	MediaId primitive.ObjectID `json:"-" bson:"_id"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId types.ContentId `json:"mediaId" bson:"mediaId"`

	// Slices of files whos content hash to the contentId
	FileIds []types.FileId `json:"fileIds" bson:"fileIds"`

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	FullresCacheIds []types.FileId `json:"fullresCacheIds" bson:"fullresCacheIds"`

	// WEBP thumbnail cache fileId
	ThumbnailCacheId types.FileId `json:"thumbnailCacheId" bson:"thumbnailCacheId"`

	CreateDate time.Time `json:"createDate" bson:"createDate"`

	// User who owns the file that resulted in this media being created
	Owner types.User `json:"owner" bson:"owner"`

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

	/* NON-DATABASE FIELDS */

	// Real media type of the media, loaded from the MimeType
	mediaType types.MediaType

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
	thumbCacheFile    types.WeblensFile
	fullresCacheFiles []types.WeblensFile

	streamer *VideoStreamer
}

func New(contentId types.ContentId) types.Media {
	return &Media{
		ContentId: contentId,
	}
}

func (m *Media) LoadFromFile(f types.WeblensFile, preReadBytes []byte, task types.Task) (retM types.Media, err error) {
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
			util.ErrTrace(err)
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

func (m *Media) ID() types.ContentId {
	return m.ContentId
}

func (m *Media) SetContentId(id types.ContentId) {
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

	return types.NewWeblensError("Not impl")
}

func (m *Media) GetMediaType() types.MediaType {
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
	return m.VideoLength
}

func (m *Media) SetOwner(owner types.User) {
	m.Owner = owner
}

func (m *Media) GetOwner() types.User {
	return m.Owner
}

func (m *Media) ReadDisplayable(q types.Quality, index int) (data []byte, err error) {
	return types.SERV.MediaRepo.FetchCacheImg(m, q, index)
}

func (m *Media) GetFiles() []types.FileId {
	return m.FileIds
}

func (m *Media) getExistingFiles() []types.FileId {
	return util.Filter(
		m.FileIds, func(fId types.FileId) bool {
			return types.SERV.FileTree.Get(fId) != nil
		},
	)
}

func (m *Media) AddFile(f types.WeblensFile) error {
	m.FileIds = util.AddToSet(m.FileIds, []types.FileId{f.ID()})
	if m.IsImported() {
		return types.SERV.StoreService.AddFileToMedia(m.ID(), f.ID())
	}
	return nil
}

func (m *Media) RemoveFile(f types.WeblensFile) error {
	var existed bool
	m.FileIds, _, existed = util.YoinkFunc(m.FileIds, func(existFile types.FileId) bool { return existFile == f.ID() })

	if !existed {
		util.Warning.Println("Attempted to remove file from Media that did not have that file")
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

func (m *Media) IsCached() bool {
	if m.thumbCacheFile == nil {
		if m.ThumbnailCacheId == "" {
			return false
		}
		cache := types.SERV.FileTree.Get(m.ThumbnailCacheId)
		if cache == nil {
			return false
		}
		m.thumbCacheFile = cache
	}

	if len(m.fullresCacheFiles) == 0 {
		m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
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
	thumbBytes, err := m.ReadDisplayable(types.Thumbnail, 0)
	if err != nil {
		return
	}

	i, err = webp.Decode(bytes.NewBuffer(thumbBytes), nil)
	if err != nil {
		return
	}

	promColors, err := prominentcolor.Kmeans(i)
	prom = util.Map(promColors, func(p prominentcolor.ColorItem) string { return p.AsString() })
	return
}

// Private

func (m *Media) loadExif(f types.WeblensFile) error {
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

func (m *Media) parseExif(f types.WeblensFile) error {

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
		m.mediaType = types.SERV.MediaRepo.TypeService().ParseMime(mimeType)

		if m.mediaType == nil {
			err = types.NewWeblensError(fmt.Sprintln("failed to parse media type:", mimeType))
			return err
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
				return types.NewWeblensError("invalid movie format")
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
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
	}

	if len(m.fullresCacheFiles) != m.PageCount {
		m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
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
			util.Debug.Println("empty orientation")
		default:
			err = types.NewWeblensError(fmt.Sprintf("Unknown rotate name [%s]", m.rotate))
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
		return types.NewWeblensError("cannot load multi-page image without page count")
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

func (m *Media) GetCacheFile(q types.Quality, generateIfMissing bool, pageNum int) (f types.WeblensFile, err error) {
	if q == types.Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	} else if q == types.Fullres && len(m.fullresCacheFiles) > pageNum && m.fullresCacheFiles[pageNum] != nil {
		f = m.fullresCacheFiles[pageNum]
		return
	}

	if pageNum >= m.PageCount {
		return nil, dataStore.ErrPageOutOfRange
	}

	cacheRoot := types.SERV.FileTree.Get("CACHE")

	var cacheFileId types.FileId
	if q == types.Fullres && len(m.FullresCacheIds) > pageNum && m.FullresCacheIds[pageNum] != "" {
		cacheFileId = m.FullresCacheIds[pageNum]
	} else if q == types.Fullres {
		cacheFileId = m.getCacheId(q, pageNum, cacheRoot)
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
		m.FullresCacheIds[pageNum] = cacheFileId
	}

	if q == types.Thumbnail {
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

			util.Warning.Printf("Cache file F[%s] not found for M[%s], generating...", cacheFileId, m.ID())
			err = m.handleCacheCreation(realFile)
			if err != nil {
				return
			}

			return m.GetCacheFile(q, false, pageNum)

		} else {
			return nil, dataStore.ErrNoCache()
		}
	}

	if q == types.Thumbnail {
		m.thumbCacheFile = f
	} else if q == types.Fullres {
		m.fullresCacheFiles[pageNum] = f
	}

	return
}

const ThumbnailHeight float32 = 500

func (m *Media) handleCacheCreation(f types.WeblensFile) (err error) {
	_, err = m.GetCacheFile(types.Thumbnail, false, 0)
	hasThumbCache := err == nil
	_, err = m.GetCacheFile(types.Fullres, false, 0)
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
			util.Error.Println(errOut.String())
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
			util.ShowErr(err)
		} else {
			thumbRatio := float64(thumbSize.Width) / float64(thumbSize.Height)
			mediaRatio := float64(m.MediaWidth) / float64(m.MediaHeight)
			// util.Debug.Println(thumbRatio, mediaRatio)
			if (thumbRatio < 1 && mediaRatio > 1) || (thumbRatio > 1 && mediaRatio < 1) {
				util.Error.Println("Mismatched media sizes")
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
		m.cacheDisplayable(types.Thumbnail, thumbBytes, 0, f.GetTree())
	}

	if !hasFullresCache {
		if m.mediaType.IsMultiPage() {
			for page := range m.PageCount {
				m.cacheDisplayable(types.Fullres, m.images[page].Image(), page, f.GetTree())
			}
		} else {
			m.cacheDisplayable(types.Fullres, m.image.Image(), 0, f.GetTree())
		}
	}

	return
}

func (m *Media) cacheDisplayable(q types.Quality, data []byte, pageNum int, ft types.FileTree) types.WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

	if len(data) == 0 {
		util.ErrTrace(errors.New("no data while trying to cache displayable"))
		return nil
	}

	cacheRoot := ft.Get("CACHE")
	f, err := ft.Touch(cacheRoot, cacheFileName, false, nil)
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists()) {
		util.ErrTrace(err)
		return nil
	} else if errors.Is(err, types.ErrFileAlreadyExists()) {
		return f
	}

	err = f.Write(data)
	if err != nil {
		util.ErrTrace(err)
		return f
	}

	if len(m.FullresCacheIds) != m.PageCount {
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
	}

	if q == types.Thumbnail && m.ThumbnailCacheId == "" || m.thumbCacheFile == nil {
		m.ThumbnailCacheId = f.ID()
		m.thumbCacheFile = f
	} else if q == types.Fullres && m.FullresCacheIds[pageNum] == "" || m.fullresCacheFiles[pageNum] == nil {
		m.FullresCacheIds[pageNum] = f.ID()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *Media) getCacheId(q types.Quality, pageNum int, cacheDir types.WeblensFile) types.FileId {
	return cacheDir.GetTree().GenerateFileId(
		filepath.Join(
			cacheDir.GetAbsPath(), m.getCacheFilename(
				q,
				pageNum,
			),
		),
	)
}

func (m *Media) getCacheFilename(q types.Quality, pageNum int) string {
	var cacheFileName string

	if m.PageCount == 1 || q == types.Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.cache", m.ID(), q)
	} else if q != types.Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.cache", m.ID(), q, pageNum)
	}

	return cacheFileName
}

func (m *Media) getImageRecognitionTags() (err error) {
	return nil
	bs, err := m.ReadDisplayable(types.Thumbnail, 0)
	if err != nil {
		return
	}
	imgBuf := bytes.NewBuffer(bs)

	resp, err := http.Post(util.GetImgRecognitionUrl()+"/recognize", "application/jpeg", imgBuf)
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
		return types.WeblensErrorMsg("failed to parse contentId")
	}
	m.ContentId = types.ContentId(contentId)

	fileIds, err := raw.Lookup("fileIds").Array().Elements()
	if err != nil {
		return types.WeblensErrorFromError(err)
	}
	m.FileIds = util.Map(
		fileIds, func(e bson.RawElement) types.FileId {
			return types.FileId(e.Value().StringValue())
		},
	)

	m.ThumbnailCacheId = types.FileId(raw.Lookup("thumbnailCacheId").StringValue())

	m.PageCount = int(raw.Lookup("pageCount").Int32())

	videoLength, ok := raw.Lookup("videoLength").Int32OK()
	if ok {
		m.VideoLength = int(videoLength)
	}

	fullresIds, err := raw.Lookup("fullresCacheIds").Array().Elements()
	if err != nil {
		return types.WeblensErrorFromError(err)
	}
	m.FullresCacheIds = util.Map(
		fullresIds, func(e bson.RawElement) types.FileId {
			return types.FileId(e.Value().StringValue())
		},
	)

	m.fullresCacheFiles = util.Map(
		m.FullresCacheIds, func(fId types.FileId) types.WeblensFile {
			return types.SERV.FileTree.Get(fId)
		},
	)

	// TODO - figure out why this happens
	if len(m.fullresCacheFiles) != m.PageCount {
		util.Warning.Printf("Bad fullres file count, got %d but expected %d", len(m.fullresCacheFiles), m.PageCount)
		m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
	}

	m.BlurHash = raw.Lookup("blurHash").StringValue()
	m.Owner = types.SERV.UserService.Get(types.Username(raw.Lookup("owner").StringValue()))
	m.MediaWidth = int(raw.Lookup("width").Int32())
	m.MediaHeight = int(raw.Lookup("height").Int32())
	m.CreateDate = time.UnixMilli(raw.Lookup("createDate").Int64())
	m.MimeType = raw.Lookup("mimeType").StringValue()

	rtArr, ok := raw.Lookup("recognitionTags").ArrayOK()
	if ok {
		rts, err := rtArr.Values()
		if err != nil {
			return types.WeblensErrorFromError(err)
		}

		m.RecognitionTags = util.Map(
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
		// "blurHash":         m.BlurHash,
		"owner":      m.Owner.GetUsername(),
		"width":      m.MediaWidth,
		"height":     m.MediaHeight,
		"createDate": m.CreateDate.UnixMilli(),
		"mimeType":   m.MimeType,
		// "recognitionTags":  m.RecognitionTags,
		"pageCount": m.PageCount,
		"imported":  m.imported,
		"hidden":    m.Hidden,
		// "videoLength":      m.VideoLength,
	}

	if m.VideoLength != 0 {
		data["videoLength"] = m.VideoLength
	}

	return json.Marshal(data)
}

func (m *Media) UnmarshalJSON(bs []byte) error {
	data := map[string]any{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}

	m.ContentId = types.ContentId(data["contentId"].(string))
	m.ThumbnailCacheId = types.FileId(data["thumbnailCacheId"].(string))

	idsStrings := util.SliceConvert[string](data["fullresCacheIds"].([]any))

	m.FullresCacheIds = util.Map(
		idsStrings, func(s string) types.FileId {
			return types.FileId(s)
		},
	)
	if data["blurHash"] != nil {
		m.BlurHash = data["blurHash"].(string)
	}
	m.Owner = types.SERV.UserService.Get(types.Username(data["owner"].(string)))
	m.MediaWidth = int(data["width"].(float64))
	m.MediaHeight = int(data["height"].(float64))
	m.CreateDate = time.UnixMilli(int64(data["createDate"].(float64)))
	m.MimeType = data["mimeType"].(string)

	if data["recognitionTags"] != nil {
		m.RecognitionTags = util.SliceConvert[string](data["recognitionTags"].([]any))
	}

	m.PageCount = int(data["pageCount"].(float64))
	m.imported = data["imported"].(bool)
	m.Hidden = data["hidden"].(bool)

	if data["videoLength"] != nil {
		m.VideoLength = int(data["videoLength"].(float64))
	}

	return nil
}
