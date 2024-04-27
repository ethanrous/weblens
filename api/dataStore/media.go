package dataStore

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
)

func NewMedia(contentId types.ContentId) types.Media {
	m := MediaMapGet(contentId)
	if m == nil {
		m = &media{
			ContentId: contentId,
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

	if m.imgBytes == nil {
		if preReadBytes != nil {
			m.imgBytes = preReadBytes
		} else {
			err = m.readFileBytes(f)
			if err != nil {
				return
			}
			task.SwLap("Read File")
		}
	}

	if !slices.Contains(m.FileIds, f.Id()) {
		m.FileIds = append(m.FileIds, f.Id())
	}

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
	task.SwLap("Generate Image")
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

	return m, nil
}

func (m *media) Id() types.ContentId {
	return m.ContentId
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

func (m *media) ReadDisplayable(q types.Quality, index ...int) (data []byte, err error) {
	defer m.Clean()

	var pageNum int
	if len(index) != 0 && (index[0] != 0 && index[0] >= m.PageCount) {
		return nil, ErrPageOutOfRange
	} else if len(index) != 0 {
		pageNum = index[0]
	} else {
		pageNum = 0
	}

	var redisKey string
	if util.ShouldUseRedis() {
		redisKey = fmt.Sprintf("%s-%s_%d", m.Id(), q, pageNum)
		var redisData string
		redisData, err = fddb.RedisCacheGet(redisKey)
		if err == nil {
			data = []byte(redisData)
			return
		}
	}

	f, err := m.getCacheFile(q, true, pageNum)
	if f == nil || err != nil {
		util.ErrTrace(err)
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

	if util.ShouldUseRedis() {
		fddb.RedisCacheSet(redisKey, string(data))
	}

	return
}

func (m *media) AddFile(f types.WeblensFile) {
	m.FileIds = util.AddToSet(m.FileIds, []types.FileId{f.Id()})
	fddb.addFileToMedia(m, f)
}

func (m *media) GetFiles() []types.FileId {
	return m.FileIds
}

func (m *media) RemoveFile(fId types.FileId) {
	var existed bool
	m.FileIds, _, existed = util.YoinkFunc(m.FileIds, func(f types.FileId) bool { return f == fId })

	if !existed {
		util.Warning.Println("Attempted to remove file from media that did not have that file")
	}

	// if len(m.FileIds) == 0 {
	// 	removeMedia(m)
	// } else {
	// }
	fddb.removeFileFromMedia(m.Id(), fId)
}

// Toss the cached data for the media generated when parsing a file.
// This will drastically reduce memory usage if used properly
func (m *media) Clean() {
	m.imgBytes = nil
	m.rawExif = nil
	m.image = nil
	m.images = nil
}

func (m *media) SetImported(i bool) {
	if !m.Imported && i {
		m.Imported = true
		mediaMapAdd(m)
	} else if !i {
		m.Imported = false
	}
}

func (m *media) IsImported() bool {
	if m == nil {
		return false
	}
	return m.Imported
}

func (m *media) Save() error {
	if !m.Imported {
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

	if m.CreateDate.Unix() == 0 && m.mediaType != nil && m.imgBytes != nil {
		return nil
	}

	// We don't need the exif data once we leave this method.
	defer func() { m.rawExif = nil }()

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

	if m.rotate == "" {
		rotate := m.rawExif["Orientation"]
		if rotate != nil {
			m.rotate = rotate.(string)
		}
	}

	if !m.mediaType.IsDisplayable() || m.mediaType.rawThumbExifKey == "" {
		return nil
	}

	raw64 := m.rawExif[m.mediaType.rawThumbExifKey].(string)
	raw64 = raw64[strings.Index(raw64, ":")+1:]

	imgBytes, err := base64.StdEncoding.DecodeString(raw64)
	if err != nil {
		return err
	}
	m.imgBytes = imgBytes

	return nil
}

func (m *media) readFileBytes(f types.WeblensFile) (err error) {
	var fileBytes []byte
	fileBytes, err = f.ReadAll()
	if err != nil {
		return
	} else if len(fileBytes) == 0 {
		err = errors.New("failed to read image from file")
		return
	}
	m.imgBytes = fileBytes
	return
}

func (m *media) generateImage() (err error) {
	if len(m.imgBytes) == 0 {
		return errors.New("cannot generate media image with no imgBytes")
	}

	bi := bimg.NewImage(m.imgBytes)

	// Rotation is backwards because imaging rotates CW, but exif stores CW rotation
	switch m.rotate {
	case "Rotate 270 CW":
		_, err = bi.Rotate(270)
	case "Rotate 90 CW":
		_, err = bi.Rotate(90)
	}
	util.ErrTrace(err)
	m.image = bi

	b, err := bi.Process(bimg.Options{Type: bimg.WEBP})
	if err != nil {
		return
	}
	m.imgBytes = b

	imgSize, err := bi.Size()
	if err != nil {
		return
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	return nil
}

func (m *media) generateImages() (err error) {
	if len(m.imgBytes) == 0 {
		return errors.New("cannot generate media image with no imgBytes")
	}
	if m.PageCount == 0 {
		return errors.New("cannot load multi-page image without page count")
	}

	m.images = make([]*bimg.Image, m.PageCount)
	var bi *bimg.Image
	for page := range m.PageCount {
		bi = bimg.NewImage(m.imgBytes)
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

// func (m *Media) generateBlurhash() (err error) {
// 	return
// if m.BlurHash == "" {
// 	m.BlurHash, err = blurhash.Encode(3, 3, m.image)
// }
// }

func (m *media) getCacheFile(q types.Quality, generateIfMissing bool, pageNum int) (f types.WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	}

	if pageNum >= m.PageCount {
		return nil, ErrPageOutOfRange
	}

	var cacheFileId types.FileId
	if q == Fullres && m.fullresCacheFiles[pageNum] != nil && m.fullresCacheFiles[pageNum].Exists() {
		f = m.fullresCacheFiles[pageNum]
		return
	} else if q == Fullres {
		if m.FullresCacheIds[pageNum] == "" {
			m.FullresCacheIds[pageNum] = m.getCacheId(q, pageNum)
		}
		cacheFileId = m.FullresCacheIds[pageNum]
	} else if q == Thumbnail {
		if m.ThumbnailCacheId == "" {
			m.ThumbnailCacheId = m.getCacheId(q, pageNum)
		}
		cacheFileId = m.ThumbnailCacheId
	}

	f = FsTreeGet(cacheFileId)
	if f == nil || !f.Exists() {
		if generateIfMissing {
			realFile, err := GetRealFile(m)
			if err != nil {
				return nil, err
			}

			t := tasker.ScanFile(realFile, nil, globalCaster)
			t.Wait()
			terr := t.ReadError()
			if terr != nil {
				return nil, terr.(error)
			}
			f = FsTreeGet(cacheFileId)
			if f == nil {
				return nil, ErrNoFile
			}
		} else {
			return nil, ErrNoFile
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
	if len(m.imgBytes) == 0 {
		if m.mediaType.IsRaw() {
			err = m.parseExif(f)
			if err != nil {
				return
			}
		} else {
			err = m.readFileBytes(f)
			if err != nil {
				return
			}
		}
	}

	if m.image == nil && !m.mediaType.multiPage {
		err = m.generateImage()
		if err != nil {
			return
		}
	} else if m.mediaType.multiPage && (len(m.images) == 0 || m.images[0] == nil) {
		err = m.generateImages()
		if err != nil {
			return
		}
	}

	thumbW := int((THUMBNAIL_HEIGHT / float32(m.MediaHeight)) * float32(m.MediaWidth))

	var thumbBytes []byte
	if !m.mediaType.multiPage && m.image != nil {
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

	m.cacheDisplayable(Thumbnail, thumbBytes, 0)

	if m.mediaType.multiPage {
		for page := range m.PageCount {
			m.cacheDisplayable(Fullres, m.images[page].Image(), page)
		}
	} else {
		m.cacheDisplayable(Fullres, m.imgBytes, 0)
	}

	return
}

func (m *media) cacheDisplayable(q types.Quality, data []byte, pageNum int) types.WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

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

	if q == Fullres {
		if m.fullresCacheFiles == nil {
			m.fullresCacheFiles = make([]types.WeblensFile, m.PageCount)
		}
		if m.FullresCacheIds == nil {
			m.FullresCacheIds = make([]types.FileId, m.PageCount)
		}
	}

	if q == Thumbnail && m.ThumbnailCacheId == "" {
		m.ThumbnailCacheId = f.Id()
		m.thumbCacheFile = f
	} else if q == Fullres && m.FullresCacheIds[pageNum] == "" {
		m.FullresCacheIds[pageNum] = f.Id()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *media) getCacheId(q types.Quality, pageNum int) types.FileId {
	if q == Thumbnail && m.ThumbnailCacheId != "" {
		return m.ThumbnailCacheId
	} else if q == Fullres && m.FullresCacheIds[pageNum] != "" {
		return m.FullresCacheIds[pageNum]
	}
	absPath := filepath.Join(GetCacheDir().GetAbsPath(), m.getCacheFilename(q, pageNum))
	return generateFileId(absPath)
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
	var imgBuf *bytes.Buffer
	if m.imgBytes == nil {
		var f types.WeblensFile
		f, err = m.getCacheFile(Thumbnail, false, 0)
		if err != nil {
			return err
		}

		var osFile *os.File
		osFile, err = f.Read()
		if err != nil {
			return
		}

		var i image.Image
		i, _, err = image.Decode(osFile)
		if err != nil {
			return
		}

		err = jpeg.Encode(imgBuf, i, nil)
		if err != nil {
			return err
		}
	} else {
		imgBuf = bytes.NewBuffer(m.imgBytes)
	}

	// resp, err := http.Get(util.GetImgRecognitionUrl() + "/ping")
	// util.Debug.Println(resp.Status, err)

	resp, err := http.Post(util.GetImgRecognitionUrl()+"/recognize", "application/jpeg", imgBuf)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get recognition tags: %s", resp.Status)
	}

	// var recogTags imageRecogResp
	var recogTags []string
	json.NewDecoder(resp.Body).Decode(&recogTags)
	// json.Unmarshal(bodyBytes, &recogTags)

	// labels := util.Map(recogTags.Labels, func(i imageRecogTag) string { return i.Label })
	m.RecognitionTags = recogTags

	return
}

func (m *media) toMarshalable() marshalableMedia {
	return marshalableMedia{
		MediaId:          m.ContentId,
		FileIds:          m.FileIds,
		ThumbnailCacheId: m.ThumbnailCacheId,
		FullresCacheIds:  m.FullresCacheIds,
		BlurHash:         m.BlurHash,
		Owner:            m.Owner.GetUsername(),
		MediaWidth:       m.MediaWidth,
		MediaHeight:      m.MediaHeight,
		CreateDate:       m.CreateDate,
		MimeType:         m.MimeType,
		RecognitionTags:  m.RecognitionTags,
		PageCount:        m.PageCount,
		Hidden:           m.Hidden,
	}
}

func (m *media) MarshalBSON() ([]byte, error) {
	return bson.Marshal(m.toMarshalable())
}

// func (m *media) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(m.toMarshalable())
// }

func marshalableToMedia(m marshalableMedia) *media {
	return &media{
		ContentId:        m.MediaId,
		FileIds:          m.FileIds,
		ThumbnailCacheId: m.ThumbnailCacheId,
		FullresCacheIds:  m.FullresCacheIds,
		BlurHash:         m.BlurHash,
		Owner:            GetUser(m.Owner).(*user),
		MediaWidth:       m.MediaWidth,
		MediaHeight:      m.MediaHeight,
		CreateDate:       m.CreateDate,
		MimeType:         m.MimeType,
		RecognitionTags:  m.RecognitionTags,
		PageCount:        m.PageCount,
	}
}
