package dataStore

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Media) LoadFromFile(f types.WeblensFile, task types.Task) (media types.Media, err error) {
	err = m.parseExif(f)
	if err != nil {
		return
	}
	task.SwLap("Parse Exif")

	if m.imgBytes == nil {
		err = m.readFileBytes(f)
		if err != nil {
			return
		}
		task.SwLap("Read File")
	}

	if m.mediaId == "" {
		err = m.generateMediaId()
		if err != nil {
			return
		}
		task.SwLap("Generate MediaId")
	}

	tmpM, err := MediaMapGet(m.Id())
	if err != nil && err != ErrNoMedia {
		return
	}
	task.SwLap("Check if exists")

	var cacheExists bool
	if tmpM != nil {
		storedM := tmpM.(*Media)
		err = f.SetMedia(storedM)
		if err != nil {
			return
		}
		task.SwLap("Set file")

		storedM.imgBytes = m.imgBytes
		m = storedM

		m.thumbCacheFile = FsTreeGet(m.thumbnailCacheId)
		for page := range m.pageCount {
			m.fullresCacheFiles[page] = FsTreeGet(m.fullresCacheIds[page])
		}

		// Check cache files exist
		if m.thumbCacheFile != nil && m.fullresCacheFiles[0] != nil {
			cacheExists = true
		}

		task.SwLap("Check cache exists")
	}

	m.AddFile(f)
	m.owner = f.Owner()
	task.SwLap("Add file and set owner")

	// if m.BlurHash == "" || !cacheExists {
	// if !cacheExists {

	// }

	// if m.BlurHash == "" {
	// 	err = m.generateBlurhash()
	// 	if err != nil {
	// 		return
	// 	}
	// 	task.SwLap("Generate blurhash")
	// }

	if !cacheExists {
		if m.mediaType.multiPage {
			err = m.generateImages()
		} else {
			err = m.generateImage()
		}
		if err != nil {
			util.ErrTrace(err)
			return
		}
		task.SwLap("Generate Image")
		err = m.handleCacheCreation(f)
		if err != nil {
			return
		}
		task.SwLap("Create cache")
	}

	if m.recognitionTags == nil && m.mediaType.supportsImgRecog {
		err = m.getImageRecognitionTags()
		if err != nil {
			util.ErrTrace(err)
		}
		task.SwLap("Get img recognition tags")
	}

	return m, nil
}

func (m *Media) Id() types.MediaId {
	if m.mediaId == "" {
		if len(m.fileIds) == 0 {
			err := errors.New("trying to generate mediaId for media with no FileId")
			util.ErrTrace(err)
			return ""
		}
		err := m.generateMediaId()
		util.ErrTrace(err)
	}

	return m.mediaId
}

func (m *Media) IsFilledOut() (bool, string) {
	if m.mediaId == "" {
		return false, "mediaId"
	}
	if len(m.fileIds) == 0 {
		return false, "file id"
	}
	if m.owner == nil {
		return false, "owner"
	}
	if m.mediaType.supportsImgRecog && m.recognitionTags == nil {
		return false, "recognition tags"
	}

	// Visual media specific properties
	if m.mediaType != nil && m.mediaType.IsDisplayable() {
		// if m.BlurHash == "" {
		// 	return false, "blurhash"
		// }
		if m.mediaWidth == 0 {
			return false, "media width"
		}
		if m.mediaHeight == 0 {
			return false, "media height"
		}
		// if m.ThumbWidth == 0 {
		// 	return false, "thumb width"
		// }
		// if m.ThumbHeight == 0 {
		// 	return false, "thumb height"
		// }

	}

	if m.createDate.IsZero() {
		return false, "create date"
	}

	return true, ""
}

func (m *Media) GetCreateDate() time.Time {
	return m.createDate
}

func (m *Media) GetMediaType() types.MediaType {
	return m.mediaType
}

func (m *Media) ReadDisplayable(q types.Quality, index ...int) (data []byte, err error) {
	defer m.Clean()

	var pageNum int
	if len(index) != 0 && (index[0] != 0 && index[0] >= m.pageCount) {
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

func (m *Media) AddFile(f types.WeblensFile) {
	m.fileIds = util.AddToSet(m.fileIds, []types.FileId{f.Id()})
	fddb.addFileToMedia(m, f)
}

func (m *Media) RemoveFile(fId types.FileId) {
	var existed bool
	m.fileIds, _, existed = util.YoinkFunc(m.fileIds, func(f types.FileId) bool { return f == fId })

	if !existed {
		util.Warning.Println("Attempted to remove file from media that did not have that file")
	}

	if len(m.fileIds) == 0 {
		removeMedia(m)
	} else {
		fddb.removeFileFromMedia(m.Id(), fId)
	}
}

// Toss the cached data for the media generated when parsing a file.
// This will drastically reduce memory usage if used properly
func (m *Media) Clean() {
	m.imgBytes = nil
	m.rawExif = nil
	m.image = nil
	m.images = nil
}

func (m *Media) SetImported(i bool) {
	if !m.imported && i {
		m.imported = true
		mediaMapAdd(m)
	} else if !i {
		m.imported = false
	}
}

func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}
	return m.imported
}

func (m *Media) Save() error {
	if !m.imported {
		return fddb.AddMedia(m)
	} else {
		return fddb.UpdateMedia(m)
	}
}

func (m *Media) getProminentColors() (prom []string, err error) {
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

func (m *Media) loadExif(f types.WeblensFile) error {
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

func (m *Media) parseExif(f types.WeblensFile) error {

	if m.createDate.Unix() == 0 && m.mediaType != nil && m.imgBytes != nil {
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
	if m.createDate.Unix() <= 0 {
		r, ok := m.rawExif["SubSecCreateDate"]
		if !ok {
			r, ok = m.rawExif["MediaCreateDate"]
		}
		if ok {
			m.createDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
			if err != nil {
				m.createDate, err = time.Parse("2006:01:02 15:04:05.00-07:00", r.(string))
			}
			if err != nil {
				m.createDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
			}
		} else {
			m.createDate = f.ModTime()
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
		m.mimeType = mimeType
		m.mediaType = ParseMimeType(mimeType)
	}

	if m.mediaType.FriendlyName() == "PDF" {
		m.pageCount = int(m.rawExif["PageCount"].(float64))
	} else {
		m.pageCount = 1
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

func (m *Media) readFileBytes(f types.WeblensFile) (err error) {
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

func (m *Media) generateImage() (err error) {
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

	m.mediaHeight = imgSize.Height
	m.mediaWidth = imgSize.Width

	return nil
}

func (m *Media) generateImages() (err error) {
	if len(m.imgBytes) == 0 {
		return errors.New("cannot generate media image with no imgBytes")
	}
	if m.pageCount == 0 {
		return errors.New("cannot load multipage image without page count")
	}

	m.images = make([]*bimg.Image, m.pageCount)
	var bi *bimg.Image
	for page := range m.pageCount {
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

	m.mediaHeight = imgSize.Height
	m.mediaWidth = imgSize.Width

	return nil
}

func (m *Media) generateMediaId() (err error) {
	if m.imgBytes == nil {
		return errors.New("cannot generate media mediaId with no imgBytes")
	}

	h := sha256.New()

	_, err = io.Copy(h, bytes.NewReader(m.imgBytes))
	if err != nil {
		util.ErrTrace(err)
		return
	}

	m.mediaId = types.MediaId(base64.URLEncoding.EncodeToString(h.Sum(nil))[:8])
	return
}

// func (m *Media) generateBlurhash() (err error) {
// 	return
// if m.BlurHash == "" {
// 	m.BlurHash, err = blurhash.Encode(3, 3, m.image)
// }
// }

func (m *Media) getCacheFile(q types.Quality, generateIfMissing bool, pageNum int) (f types.WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	}

	if pageNum >= m.pageCount {
		return nil, ErrPageOutOfRange
	}

	var cacheFileId types.FileId
	if q == Fullres && m.fullresCacheFiles[pageNum] != nil && m.fullresCacheFiles[pageNum].Exists() {
		f = m.fullresCacheFiles[pageNum]
		return
	} else if q == Fullres {
		if m.fullresCacheIds[pageNum] == "" {
			m.fullresCacheIds[pageNum] = m.getCacheId(q, pageNum)
		}
		cacheFileId = m.fullresCacheIds[pageNum]
	} else if q == Thumbnail {
		if m.thumbnailCacheId == "" {
			m.thumbnailCacheId = m.getCacheId(q, pageNum)
		}
		cacheFileId = m.thumbnailCacheId
	}

	f = FsTreeGet(cacheFileId)
	if f == nil || !f.Exists() {
		if generateIfMissing {
			realFile, err := GetRealFile(m)
			if err != nil {
				return nil, err
			}

			t := tasker.ScanFile(realFile, m, globalCaster)
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

func (m *Media) handleCacheCreation(f types.WeblensFile) (err error) {
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

	thumbW := int((THUMBNAIL_HEIGHT / float32(m.mediaHeight)) * float32(m.mediaWidth))

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
		for page := range m.pageCount {
			m.cacheDisplayable(Fullres, m.images[page].Image(), page)
		}
	} else {
		m.cacheDisplayable(Fullres, m.imgBytes, 0)
	}

	return
}

func (m *Media) cacheDisplayable(q types.Quality, data []byte, pageNum int) types.WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

	f, err := Touch(GetCacheDir(), cacheFileName, true)
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
			m.fullresCacheFiles = make([]types.WeblensFile, m.pageCount)
		}
		if m.fullresCacheIds == nil {
			m.fullresCacheIds = make([]types.FileId, m.pageCount)
		}
	}

	if q == Thumbnail && m.thumbnailCacheId == "" {
		m.thumbnailCacheId = f.Id()
		m.thumbCacheFile = f
	} else if q == Fullres && m.fullresCacheIds[pageNum] == "" {
		m.fullresCacheIds[pageNum] = f.Id()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *Media) getCacheId(q types.Quality, pageNum int) types.FileId {
	if q == Thumbnail && m.thumbnailCacheId != "" {
		return m.thumbnailCacheId
	} else if q == Fullres && m.fullresCacheIds[pageNum] != "" {
		return m.fullresCacheIds[pageNum]
	}
	absPath := filepath.Join(GetCacheDir().GetAbsPath(), m.getCacheFilename(q, pageNum))
	return types.FileId(util.GlobbyHash(8, GuaranteeRelativePath(absPath)))
}

func (m *Media) getCacheFilename(q types.Quality, pageNum int) string {
	var cacheFileName string

	if m.pageCount == 1 || q == Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.wlcache", m.Id(), q)
	} else if q != Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.wlcache", m.Id(), q, pageNum)
	}

	return cacheFileName
}

func (m *Media) getImageRecognitionTags() (err error) {
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

	resp, err := http.Post(util.GetImgRecognitionUrl(), "application/jpeg", imgBuf)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get recognition tags: %s", resp.Status)
	}

	// var regocTags imageRegocResp
	var regocTags []string
	json.NewDecoder(resp.Body).Decode(&regocTags)
	// json.Unmarshal(bodyBytes, &regocTags)

	// labels := util.Map(regocTags.Labels, func(i imageRegocTag) string { return i.Label })
	m.recognitionTags = regocTags

	return
}

func (m *Media) toMarshalable() marshalableMedia {
	return marshalableMedia{
		MediaId:          m.mediaId,
		FileIds:          m.fileIds,
		ThumbnailCacheId: m.thumbnailCacheId,
		FullresCacheIds:  m.fullresCacheIds,
		BlurHash:         m.blurHash,
		Owner:            m.owner.GetUsername(),
		MediaWidth:       m.mediaWidth,
		MediaHeight:      m.mediaHeight,
		CreateDate:       m.createDate,
		MimeType:         m.mimeType,
		RecognitionTags:  m.recognitionTags,
		PageCount:        m.pageCount,
	}
}

func (m *Media) MarshalBSON() ([]byte, error) {
	return bson.Marshal(m.toMarshalable())
}

func (m *Media) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.toMarshalable())
}

func marshalableToMedia(m marshalableMedia) *Media {
	return &Media{
		mediaId:          m.MediaId,
		fileIds:          m.FileIds,
		thumbnailCacheId: m.ThumbnailCacheId,
		fullresCacheIds:  m.FullresCacheIds,
		blurHash:         m.BlurHash,
		owner:            GetUser(m.Owner),
		mediaWidth:       m.MediaWidth,
		mediaHeight:      m.MediaHeight,
		createDate:       m.CreateDate,
		mimeType:         m.MimeType,
		recognitionTags:  m.RecognitionTags,
		pageCount:        m.PageCount,
	}
}
