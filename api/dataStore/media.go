package dataStore

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewMedia(contentId types.ContentId) types.Media {
	em := MediaMapGet(contentId)
	var m *media
	if em == nil {
		m = &media{
			ContentId: contentId,
		}
	} else {
		realE := em.(*media)
		m = &media{
			ContentId:       contentId,
			RecognitionTags: realE.RecognitionTags,
			imported:        true,
			Owner:           realE.Owner,
		}
	}

	return m
}

func (m *media) LoadFromFile(f types.WeblensFile, preReadBytes []byte, task types.Task) (retM types.Media, err error) {
	err = m.parseExif(f)
	if err != nil {
		return
	}
	task.SwLap("Parse Exif")

	err = m.handleCacheCreation(f)
	if err != nil {
		return
	}
	task.SwLap("Create cache")

	if m.RecognitionTags == nil && m.mediaType.supportsImgRecog {
		err = m.getImageRecognitionTags()
		if err != nil {
			util.ErrTrace(err)
		}
		task.SwLap("Get img recognition tags")
	}

	m.AddFile(f)

	m.SetOwner(f.Owner())
	task.SwLap("Add file and set owner")

	// if m.MediaType.multiPage {
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

func (m *media) Id() types.ContentId {
	return m.ContentId
}

func (m *media) SetContentId(id types.ContentId) {
	m.ContentId = id
}

func (m *media) IsFilledOut() (bool, string) {
	if m.ContentId == "" {
		return false, "mediaId"
	}
	if len(m.FileIds) == 0 {
		return false, "file id"
	}
	if m.Owner == nil {
		return false, "owner"
	}
	if m.mediaType.supportsImgRecog && m.RecognitionTags == nil {
		return false, "recognition tags"
	}

	// Visual media specific properties
	if m.mediaType != nil && m.mediaType.IsDisplayable() {
		// if m.BlurHash == "" {
		// 	return false, "blurhash"
		// }
		if m.MediaWidth == 0 {
			return false, "media width"
		}
		if m.MediaHeight == 0 {
			return false, "media height"
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

func (m *media) IsHidden() bool {
	return m.Hidden
}

func (m *media) GetCreateDate() time.Time {
	return m.CreateDate
}

func (m *media) SetCreateDate(t time.Time) error {
	err := fddb.adjustMediaDate(m, t)
	if err != nil {
		return err
	}
	m.CreateDate = t

	return nil
}

func (m *media) GetMediaType() types.MediaType {
	if m.mediaType != nil && m.mediaType.mimeType != "" {
		return m.mediaType
	}
	if m.MimeType != "" {
		m.mediaType = ParseMimeType(m.MimeType)
	}
	return m.mediaType
}

func (m *media) GetPageCount() int {
	return m.PageCount
}

func (m *media) SetOwner(owner types.User) {
	m.Owner = owner.(*user)
}

func (m *media) GetOwner() types.User {
	return m.Owner
}

var thumbMap map[string][]byte = map[string][]byte{}
var thumbMapLock sync.RWMutex = sync.RWMutex{}

func (m *media) ReadDisplayable(q types.Quality, index ...int) (data []byte, err error) {
	var pageNum int
	if len(index) != 0 && (index[0] != 0 && index[0] >= m.PageCount) {
		return nil, ErrPageOutOfRange
	} else if len(index) != 0 {
		pageNum = index[0]
	} else {
		pageNum = 0
	}

	var ok bool
	cacheKey := string(m.Id()) + string(q)
	thumbMapLock.RLock()
	if data, ok = thumbMap[cacheKey]; ok {
		thumbMapLock.RUnlock()
		return
	}
	thumbMapLock.RUnlock()

	//var redisKey string
	//if util.ShouldUseRedis() {
	//	redisKey = fmt.Sprintf("%s-%s_%d", m.Id(), q, pageNum)
	//	var redisData string
	//	start := time.Now()
	//	redisData, err = fddb.RedisCacheGet(redisKey)
	//	util.Debug.Println("Redis cache get time:", time.Since(start))
	//	if err == nil {
	//		data = []byte(redisData)
	//		return
	//	}
	//}

	f, err := m.getCacheFile(q, true, pageNum)
	if err != nil {
		return
	} else if f == nil {
		return nil, ErrNoFile
	}

	data, err = f.ReadAll()
	if err != nil {
		return
	}
	if len(data) == 0 {
		err = fmt.Errorf("displayable bytes empty")
		return
	}

	//if util.ShouldUseRedis() {
	//	err = fddb.RedisCacheSet(redisKey, string(data))
	//	if err != nil {
	//		return
	//	}
	//}

	thumbMapLock.Lock()
	thumbMap[cacheKey] = data
	thumbMapLock.Unlock()

	return
}

func (m *media) GetFiles() []types.FileId {
	return m.FileIds
}

func (m *media) AddFile(f types.WeblensFile) {
	m.FileIds = util.AddToSet(m.FileIds, []types.FileId{f.Id()})
	err := fddb.addFileToMedia(m, f)
	util.ErrTrace(err)
}

func (m *media) RemoveFile(fId types.FileId) {
	var existed bool
	m.FileIds, _, existed = util.YoinkFunc(m.FileIds, func(f types.FileId) bool { return f == fId })

	if !existed {
		util.Warning.Println("Attempted to remove file from media that did not have that file")
	}

	err := fddb.removeFileFromMedia(m.Id(), fId)
	if err != nil {
		util.ErrTrace(err)
	}
	if len(m.FileIds) == 0 {
		removeMedia(m)
	}
}

// Clean Toss the cached data for the media generated when parsing a file.
// This will drastically reduce memory usage if used properly
func (m *media) Clean() {
	if m == nil {
		return
	}
	// return

	m.rawExif = nil
	m.image = nil
	m.images = nil
}

func (m *media) SetImported(i bool) {
	if !m.imported && i {
		m.imported = true
		mediaMapAdd(m)
	} else if !i {
		m.imported = false
	}
}

func (m *media) IsImported() bool {
	if m == nil {
		return false
	}

	return m.imported
}

func (m *media) IsCached() bool {
	if m.thumbCacheFile == nil {
		if m.ThumbnailCacheId == "" {
			return false
		}
		cache := FsTreeGet(m.ThumbnailCacheId)
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
				cache := FsTreeGet(m.FullresCacheIds[i])
				if cache == nil {
					return false
				}
				m.fullresCacheFiles[i] = cache
			}
		}
	}

	return true
}

func (m *media) SetEnabled(e bool) {
	m.Enabled = e
}

func (m *media) IsEnabled() bool {
	return m.Enabled
}

func (m *media) Save() error {
	if !m.imported {
		return fddb.AddMedia(m)
	} else {
		return fddb.UpdateMedia(m)
	}
}

func (m *media) getProminentColors() (prom []string, err error) {
	var i image.Image
	thumbBytes, err := m.ReadDisplayable(Thumbnail)
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

func (m *media) loadExif(f types.WeblensFile) error {
	if gexift == nil {
		err := errors.New("exiftool not initialized")
		return err
	}
	fileInfos := gexift.ExtractMetadata(f.GetAbsPath())
	if fileInfos[0].Err != nil {
		return fileInfos[0].Err
	}

	m.rawExif = fileInfos[0].Fields
	return nil
}

func (m *media) parseExif(f types.WeblensFile) error {

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
		} else {
			m.CreateDate = f.ModTime()
		}
		if err != nil {
			return err
		}
	}

	if m.mediaType == nil {
		mimeType, ok := m.rawExif["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MimeType = mimeType
		m.mediaType = ParseMimeType(mimeType)
	}

	if m.mediaType.FriendlyName() == "PDF" {
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

	if !m.mediaType.IsDisplayable() || m.mediaType.rawThumbExifKey == "" {
		return nil
	}

	return nil
}

func (m *media) generateImage(bs []byte) (err error) {
	bi := bimg.NewImage(bs)

	// Rotation is backwards because imaging rotates CW, but exif stores CW rotation
	switch m.rotate {
	case "Rotate 270 CW":
		_, err = bi.Rotate(270)
	case "Rotate 90 CW":
		_, err = bi.Rotate(90)
	}
	util.ShowErr(err)

	webpBs, err := bi.Convert(bimg.WEBP)
	util.ShowErr(err)
	if err != nil {
		return
	}

	m.image = bimg.NewImage(webpBs)

	imgSize, err := m.image.Size()
	if err != nil {
		return
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	return nil
}

func (m *media) generateImages(bs []byte) (err error) {
	if m.PageCount < 2 {
		return errors.New("cannot load multi-page image without page count")
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

func (m *media) getCacheFile(q types.Quality, generateIfMissing bool, pageNum int) (f types.WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	} else if q == Fullres && len(m.fullresCacheFiles) > pageNum && m.fullresCacheFiles[pageNum] != nil {
		f = m.fullresCacheFiles[pageNum]
		return
	}

	if pageNum >= m.PageCount {
		return nil, ErrPageOutOfRange
	}

	var cacheFileId types.FileId
	if q == Fullres && len(m.FullresCacheIds) > pageNum && m.FullresCacheIds[pageNum] != "" {
		cacheFileId = m.FullresCacheIds[pageNum]
	} else if q == Fullres {
		cacheFileId = m.getCacheId(q, pageNum)
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
		m.FullresCacheIds[pageNum] = cacheFileId
	}

	if q == Thumbnail {
		if m.ThumbnailCacheId == "" {
			m.ThumbnailCacheId = m.getCacheId(q, pageNum)
		}
		cacheFileId = m.ThumbnailCacheId
	}

	f = FsTreeGet(cacheFileId)
	if f == nil || !f.Exists() {
		if generateIfMissing && !generateIfMissing {
			realFile, err := GetRealFile(m)
			if err != nil {
				return nil, err
			}

			err = m.handleCacheCreation(realFile)
			if err != nil {
				return nil, err
			}

			return m.getCacheFile(q, false, pageNum)

		} else {
			return nil, ErrNoCache
		}
	}

	if q == Thumbnail {
		m.thumbCacheFile = f
	} else if q == Fullres {
		m.fullresCacheFiles[pageNum] = f
	}

	return
}

const THUMBNAIL_HEIGHT float32 = 500

func (m *media) handleCacheCreation(f types.WeblensFile) (err error) {
	_, err = m.getCacheFile(Thumbnail, false, 0)
	hasThumbCache := err == nil
	_, err = m.getCacheFile(Fullres, false, 0)
	hasFullresCache := err == nil

	if hasThumbCache && hasFullresCache && m.MediaWidth != 0 && m.MediaHeight != 0 {
		return nil
	}

	var bs []byte

	if m.mediaType.IsRaw() {
		raw64 := m.rawExif[m.mediaType.rawThumbExifKey].(string)
		raw64 = raw64[strings.Index(raw64, ":")+1:]

		imgBytes, err := base64.StdEncoding.DecodeString(raw64)
		if err != nil {
			return err
		}
		bs = imgBytes
	} else {
		bs, err = f.ReadAll()
		if err != nil {
			return err
		}
	}

	if m.image == nil && !m.mediaType.multiPage {
		err = m.generateImage(bs)
		if err != nil {
			return
		}
	} else if m.mediaType.multiPage && (len(m.images) == 0 || m.images[0] == nil) {
		err = m.generateImages(bs)
		if err != nil {
			return
		}
	}

	if !m.mediaType.multiPage && m.image == nil || m.mediaType.multiPage && m.images == nil {
		return ErrNoImage
	}

	thumbW := int((THUMBNAIL_HEIGHT / float32(m.MediaHeight)) * float32(m.MediaWidth))

	var thumbBytes []byte
	if !m.mediaType.multiPage && m.image != nil {

		// Copy image buffer for the thumbnail
		thumbImg := bimg.NewImage(m.image.Image())

		thumbBytes, err = thumbImg.Resize(thumbW, int(THUMBNAIL_HEIGHT))
		if err != nil {
			return
		}
	}

	if m.mediaType.multiPage && len(m.images) != 0 && m.images[0] != nil {
		thumbImg := bimg.NewImage(m.images[0].Image())

		thumbBytes, err = thumbImg.Resize(thumbW, int(THUMBNAIL_HEIGHT))
		if err != nil {
			return
		}
	}

	if !hasThumbCache {
		m.cacheDisplayable(Thumbnail, thumbBytes, 0)
	}

	if !hasFullresCache {
		if m.mediaType.multiPage {
			for page := range m.PageCount {
				m.cacheDisplayable(Fullres, m.images[page].Image(), page)
			}
		} else {
			m.cacheDisplayable(Fullres, m.image.Image(), 0)
		}
	}

	return
}

func (m *media) cacheDisplayable(q types.Quality, data []byte, pageNum int) types.WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

	if len(data) == 0 {
		util.ErrTrace(errors.New("no data while trying to cache displayable"))
		return nil
	}

	f, err := Touch(GetCacheDir(), cacheFileName, false)
	if err != nil && err != ErrFileAlreadyExists {
		util.ErrTrace(err)
		return nil
	} else if err == ErrFileAlreadyExists {
		return f
	}

	err = f.Write(data)
	if err != nil {
		util.ErrTrace(err)
		return f
	}

	if q == Thumbnail && m.ThumbnailCacheId == "" || m.thumbCacheFile == nil {
		m.ThumbnailCacheId = f.Id()
		m.thumbCacheFile = f
	} else if q == Fullres && m.FullresCacheIds[pageNum] == "" || m.fullresCacheFiles[pageNum] == nil {
		m.FullresCacheIds[pageNum] = f.Id()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *media) getCacheId(q types.Quality, pageNum int) types.FileId {
	return generateFileId(filepath.Join(GetCacheDir().GetAbsPath(), m.getCacheFilename(q, pageNum)))
}

func (m *media) getCacheFilename(q types.Quality, pageNum int) string {
	var cacheFileName string

	if m.PageCount == 1 || q == Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.wlcache", m.Id(), q)
	} else if q != Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.wlcache", m.Id(), q, pageNum)
	}

	return cacheFileName
}

func (m *media) getImageRecognitionTags() (err error) {
	bs, err := m.ReadDisplayable(Thumbnail)
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
	json.NewDecoder(resp.Body).Decode(&recogTags)

	m.RecognitionTags = recogTags

	return
}

func (m *media) MarshalBSON() ([]byte, error) {
	data := map[string]any{
		"contentId":        m.ContentId,
		"fileIds":          m.FileIds,
		"thumbnailCacheId": m.ThumbnailCacheId,
		"fullresCacheIds":  m.FullresCacheIds,
		"blurHash":         m.BlurHash,
		"owner":            m.Owner.Username,
		"width":            m.MediaWidth,
		"height":           m.MediaHeight,
		"createDate":       m.CreateDate,
		"mimeType":         m.MimeType,
		"recognitionTags":  m.RecognitionTags,
		"pageCount":        m.PageCount,
	}
	return bson.Marshal(data)
}

func (m *media) UnmarshalBSON(bs []byte) error {
	data := map[string]any{}
	err := bson.Unmarshal(bs, data)
	if err != nil {
		return err
	}

	if data["contentId"] != nil {
		m.ContentId = types.ContentId(data["contentId"].(string))
	} else if data["mediaId"] != nil {
		m.ContentId = types.ContentId(data["mediaId"].(string))
	} else {
		return errors.New("no contentId while decoding media")
	}

	m.FileIds = util.Map(data["fileIds"].(primitive.A), func(a any) types.FileId { return types.FileId(a.(string)) })
	m.ThumbnailCacheId = types.FileId(data["thumbnailCacheId"].(string))

	if data["fullresCacheIds"] != nil {
		m.FullresCacheIds = util.Map(data["fullresCacheIds"].(primitive.A), func(a any) types.FileId { return types.FileId(a.(string)) })
	}

	m.BlurHash = data["blurHash"].(string)
	m.Owner = GetUser(types.Username(data["owner"].(string))).(*user)
	m.MediaWidth = int(data["width"].(int32))
	m.MediaHeight = int(data["height"].(int32))
	m.CreateDate = time.UnixMilli(int64(data["createDate"].(primitive.DateTime)))
	m.MimeType = data["mimeType"].(string)

	if data["recognitionTags"] != nil {
		m.RecognitionTags = util.SliceConvert[string](data["recognitionTags"].(primitive.A))
	}

	m.PageCount = int(data["pageCount"].(int32))

	if data["hidden"] != nil {
		m.Hidden = data["hidden"].(bool)
	}

	return nil
}
