package models

import (
	"bytes"
	"encoding/json"
	"errors"
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

	// Full-res image dimentions
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
	// things like cache, dimentions, etc., as it might not have them.
	imported bool

	// WEBP thumbnail cache fileId
	lowresCacheFile fileTree.WeblensFile

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	highResCacheFiles []fileTree.WeblensFile
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

// func (m *Media) IsFilledOut() (bool, string) {
// 	if m.ContentId == "" {
// 		return false, "mediaId"
// 	}
// 	if len(m.FileIds) == 0 {
// 		return false, "file id"
// 	}
// 	if m.Owner == nil {
// 		return false, "owner"
// 	}
// 	if m.mediaType.SupportsImgRecog() && m.RecognitionTags == nil {
// 		return false, "recognition tags"
// 	}
//
// 	// Visual Media specific properties
// 	if m.mediaType != nil && m.mediaType.IsDisplayable() {
// 		// if m.BlurHash == "" {
// 		// 	return false, "blurhash"
// 		// }
// 		if m.Width == 0 {
// 			return false, "Media width"
// 		}
// 		if m.Height == 0 {
// 			return false, "Media height"
// 		}
// 		// if m.ThumbWidth == 0 {
// 		// 	return false, "thumb width"
// 		// }
// 		// if m.ThumbHeight == 0 {
// 		// 	return false, "thumb height"
// 		// }
//
// 	}
//
// 	if m.CreateDate.IsZero() {
// 		return false, "create date"
// 	}
//
// 	return true, ""
// }

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

func (m *Media) SetLowresCacheFile(thumb fileTree.WeblensFile) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.lowresCacheFile = thumb
}

func (m *Media) GetLowresCacheFile() fileTree.WeblensFile {
	return m.lowresCacheFile
}

func (m *Media) SetHighresCacheFiles(highresFiles []fileTree.WeblensFile) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.highResCacheFiles = highresFiles
}

func (m *Media) GetHighresCacheFiles() []fileTree.WeblensFile {
	return m.highResCacheFiles
}

// Private

// func (m *Media) generateImage(bs []byte) (err error) {
// 	m.image = bimg.NewImage(bs)
//
// 	if m.Rotate == "" {
// 		panic("AHH FILL ROTATE IN")
// 	}
//
// 	// Rotation is backwards because bimg rotates CW, but exif stores CW rotation
// 	if m.GetMediaType().IsRaw() {
// 		switch m.Rotate {
// 		case "Rotate 270 CW":
// 			_, err = m.image.Rotate(270)
// 		case "Rotate 90 CW":
// 			_, err = m.image.Rotate(90)
// 		case "Horizontal (normal)":
// 		case "":
// 			wlog.Debug.Println("empty orientation")
// 		default:
// 			err = errors.New(fmt.Sprintf("Unknown Rotate name [%s]", m.Rotate))
// 		}
// 		if err != nil {
// 			return
// 		}
// 	}
//
// 	_, err = m.image.Convert(bimg.WEBP)
// 	if err != nil {
// 		return
// 	}
//
// 	imgSize, err := m.image.Size()
// 	if err != nil {
// 		return
// 	}
//
// 	m.Height = imgSize.Height
// 	m.Width = imgSize.Width
//
// 	return nil
// }

// func (m *Media) generateImages(bs []byte) (err error) {
// 	if !m.mediaType.IsMultiPage() {
// 		return errors.New("cannot load multi-page image without page count")
// 	}
//
// 	m.images = make([]*bimg.Image, m.PageCount)
// 	var bi *bimg.Image
// 	for page := range m.PageCount {
// 		bi = bimg.NewImage(bs)
// 		pageBytes, err := bi.Process(bimg.Options{Type: bimg.WEBP, PageNum: page})
// 		if err != nil {
// 			return err
// 		}
// 		m.images[page] = bimg.NewImage(pageBytes)
// 	}
//
// 	imgSize, err := bi.Size()
// 	if err != nil {
// 		return
// 	}
//
// 	m.Height = imgSize.Height
// 	m.Width = imgSize.Width
//
// 	return nil
// }

// func (m *Media) GetCacheFile(q MediaQuality, generateIfMissing bool, pageNum int) (f *fileTree.WeblensFileImpl, err error) {
// 	if q == LowRes && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
// 		f = m.thumbCacheFile
// 		return
// 	} else if q == HighRes && len(m.fullresCacheFiles) > pageNum && m.fullresCacheFiles[pageNum] != nil {
// 		f = m.fullresCacheFiles[pageNum]
// 		return
// 	}
//
// 	if pageNum >= m.PageCount {
// 		return nil, dataStore.ErrPageOutOfRange
// 	}
//
// 	cacheRoot := types.SERV.FileTree.Get("CACHE")
//
// 	var cacheFileId fileTree.FileId
// 	if q == HighRes && len(m.highResCacheFiles) > pageNum && m.highResCacheFiles[pageNum] != "" {
// 		cacheFileId = m.highResCacheFiles[pageNum]
// 	} else if q == HighRes {
// 		cacheFileId = m.getCacheId(q, pageNum, cacheRoot)
// 		m.highResCacheFiles = make([]fileTree.FileId, m.PageCount)
// 		m.highResCacheFiles[pageNum] = cacheFileId
// 	}
//
// 	if q == LowRes {
// 		if m.lowresCacheFile == "" {
// 			m.lowresCacheFile = m.getCacheId(q, pageNum, cacheRoot)
// 		}
// 		cacheFileId = m.lowresCacheFile
// 	}
//
// 	f = types.SERV.FileTree.Get(cacheFileId)
// 	if f == nil || !f.Exists() {
// 		if generateIfMissing {
// 			realFile := types.SERV.FileTree.Get(m.getExistingFiles()[0])
// 			if realFile == nil {
// 				return nil, dataStore.ErrNoCache()
// 			}
//
// 			wlog.Warning.Printf("Cache file F[%s] not found for M[%s], generating...", cacheFileId, m.ID())
// 			err = m.handleCacheCreation(realFile)
// 			if err != nil {
// 				return
// 			}
//
// 			return m.GetCacheFile(q, false, pageNum)
//
// 		} else {
// 			return nil, dataStore.ErrNoCache()
// 		}
// 	}
//
// 	if q == LowRes {
// 		m.thumbCacheFile = f
// 	} else if q == HighRes {
// 		m.fullresCacheFiles[pageNum] = f
// 	}
//
// 	return
// }

const ThumbnailHeight float32 = 500

// func (m *Media) handleCacheCreation(f *fileTree.WeblensFileImpl) (err error) {
// 	_, err = m.GetCacheFile(LowRes, false, 0)
// 	hasThumbCache := err == nil
// 	_, err = m.GetCacheFile(HighRes, false, 0)
// 	hasFullresCache := err == nil
//
// 	if hasThumbCache && hasFullresCache && m.Width != 0 && m.Height != 0 {
// 		return nil
// 	}
//
// 	var bs []byte
//
// 	if m.mediaType.IsRaw() {
// 		if m.rawExif == nil {
// 			err = m.loadExif(f)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		raw64 := m.rawExif[m.mediaType.GetThumbExifKey()].(string)
// 		raw64 = raw64[strings.Index(raw64, ":")+1:]
//
// 		imgBytes, err := base64.StdEncoding.DecodeString(raw64)
// 		if err != nil {
// 			return err
// 		}
// 		bs = imgBytes
// 	} else if m.mediaType.IsVideo() {
// 		out := bytes.NewBuffer(nil)
// 		errOut := bytes.NewBuffer(nil)
//
// 		const frameNum = 10
//
// 		err = ffmpeg.Input(f.GetAbsPath()).Filter(
// 			"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
// 		).Output(
// 			"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
// 		).WithOutput(out).WithErrorOutput(errOut).Run()
// 		if err != nil {
// 			wlog.Error.Println(errOut.String())
// 			return err
// 		}
// 		bs = out.Bytes()
//
// 	} else {
// 		bs, err = f.ReadAll()
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	if m.image == nil && !m.mediaType.IsMultiPage() {
// 		err = m.generateImage(bs)
// 		if err != nil {
// 			return
// 		}
// 	} else if m.mediaType.IsMultiPage() && (len(m.images) == 0 || m.images[0] == nil) {
// 		err = m.generateImages(bs)
// 		if err != nil {
// 			return
// 		}
// 	}
//
// 	if !m.mediaType.IsMultiPage() && m.image == nil || m.mediaType.IsMultiPage() && m.images == nil {
// 		return dataStore.ErrNoImage
// 	}
//
// 	thumbW := int((ThumbnailHeight / float32(m.Height)) * float32(m.Width))
//
// 	var thumbBytes []byte
// 	if !m.mediaType.IsMultiPage() && m.image != nil {
//
// 		// Copy image buffer for the thumbnail
// 		thumbImg := bimg.NewImage(m.image.Image())
//
// 		thumbBytes, err = thumbImg.Resize(thumbW, int(ThumbnailHeight))
// 		if err != nil {
// 			return
// 		}
// 		thumbSize, err := thumbImg.Size()
// 		if err != nil {
// 			wlog.ShowErr(err)
// 		} else {
// 			thumbRatio := float64(thumbSize.Width) / float64(thumbSize.Height)
// 			mediaRatio := float64(m.Width) / float64(m.Height)
// 			// util.Debug.Println(thumbRatio, mediaRatio)
// 			if (thumbRatio < 1 && mediaRatio > 1) || (thumbRatio > 1 && mediaRatio < 1) {
// 				wlog.Error.Println("Mismatched media sizes")
// 			}
// 		}
// 	}
//
// 	if m.mediaType.IsMultiPage() && len(m.images) != 0 && m.images[0] != nil {
// 		thumbImg := bimg.NewImage(m.images[0].Image())
//
// 		thumbBytes, err = thumbImg.Resize(thumbW, int(ThumbnailHeight))
// 		if err != nil {
// 			return
// 		}
// 	}
//
// 	if !hasThumbCache {
// 		m.cacheDisplayable(LowRes, thumbBytes, 0, f.GetTree())
// 	}
//
// 	if !hasFullresCache {
// 		if m.mediaType.IsMultiPage() {
// 			for page := range m.PageCount {
// 				m.cacheDisplayable(HighRes, m.images[page].Image(), page, f.GetTree())
// 			}
// 		} else {
// 			m.cacheDisplayable(HighRes, m.image.Image(), 0, f.GetTree())
// 		}
// 	}
//
// 	return
// }

// func (m *Media) cacheDisplayable(q MediaQuality, data []byte, pageNum int, ft types.FileTree) *fileTree.WeblensFileImpl {
// 	cacheFileName := m.getCacheFilename(q, pageNum)
//
// 	if len(data) == 0 {
// 		wlog.ErrTrace(errors.New("no data while trying to cache displayable"))
// 		return nil
// 	}
//
// 	cacheRoot := ft.Get("CACHE")
// 	f, err := ft.Touch(cacheRoot, cacheFileName, false, nil)
// 	if err != nil {
// 		if !strings.Contains(err.Error(), "file already exists") {
// 			wlog.ErrTrace(err)
// 			return nil
// 		} else {
// 			return f
// 		}
// 	}
//
// 	err = f.Write(data)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return f
// 	}
//
// 	if len(m.highResCacheFiles) != m.PageCount {
// 		m.highResCacheFiles = make([]fileTree.FileId, m.PageCount)
// 	}
//
// 	if q == LowRes && m.lowresCacheFile == "" || m.thumbCacheFile == nil {
// 		m.lowresCacheFile = f.ID()
// 		m.thumbCacheFile = f
// 	} else if q == HighRes && m.highResCacheFiles[pageNum] == "" || m.fullresCacheFiles[pageNum] == nil {
// 		m.highResCacheFiles[pageNum] = f.ID()
// 		m.fullresCacheFiles[pageNum] = f
// 	}
//
// 	return f
// }

// func (m *Media) getCacheId(q MediaQuality, pageNum int, cacheDir *fileTree.WeblensFileImpl) fileTree.FileId {
// 	return cacheDir.GetTree().GenerateFileId(
// 		filepath.Join(
// 			cacheDir.GetAbsPath(), m.getCacheFilename(
// 				q,
// 				pageNum,
// 			),
// 		),
// 	)
// }

// func (m *Media) getCacheFilename(q MediaQuality, pageNum int) string {
// 	var cacheFileName string
//
// 	if m.PageCount == 1 || q == LowRes {
// 		cacheFileName = fmt.Sprintf("%s-%s.cache", m.ID(), q)
// 	} else if q != LowRes {
// 		cacheFileName = fmt.Sprintf("%s-%s_%d.cache", m.ID(), q, pageNum)
// 	}
//
// 	return cacheFileName
// }

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

// func (m *Media) MarshalBSON() ([]byte, error) {
// 	data := map[string]any{
// 		"contentId":        m.ContentId,
// 		"fileIds":          m.FileIds,
// 		"thumbnailCacheId": m.lowresCacheFile,
// 		"fullresCacheIds":  m.highResCacheFiles,
// 		// "blurHash":         m.BlurHash,
// 		"owner":           m.Owner,
// 		"width":           m.Width,
// 		"height":          m.Height,
// 		"createDate":      m.CreateDate,
// 		"mimeType":        m.MimeType,
// 		"recognitionTags": m.RecognitionTags,
// 		"pageCount":       m.PageCount,
// 		"videoLength":     m.Duration,
// 	}
//
// 	return bson.Marshal(data)
// }

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
		"width":  m.Width,
		"height": m.Height,
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

	// if data["blurHash"] != nil {
	// 	m.BlurHash = data["blurHash"].(string)
	// }
	// m.Owner = types.SERV.UserService.Get(Username(data["owner"].(string)))
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

	LoadMediaFromFile(m *Media, file *fileTree.WeblensFileImpl) error
	RemoveFileFromMedia(media *Media, fileId fileTree.FileId) error

	GetMediaType(m *Media) MediaType
	GetMediaTypes() MediaTypeService
	IsFileDisplayable(file *fileTree.WeblensFileImpl) bool
	IsCached(m *Media) bool

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

type ContentId string
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
			"b:v":                internal.GetVideoConstBitrate(),
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
