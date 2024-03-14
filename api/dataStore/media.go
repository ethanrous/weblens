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
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
)

func (m *Media) LoadFromFile(f *WeblensFile, task Task) (media *Media, err error) {
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

	if m.MediaId == "" {
		err = m.generateFileHash()
		if err != nil {
			return
		}
		task.SwLap("Generate Filehash")
	}

	storedM, err := MediaMapGet(m.Id())
	if err != nil && err != ErrNoMedia {
		return
	}
	task.SwLap("Check if exists")

	var cacheExists bool
	if storedM != nil {
		err = f.SetMedia(storedM)
		if err != nil {
			return
		}
		task.SwLap("Set file")

		storedM.imgBytes = m.imgBytes
		m = storedM

		m.thumbCacheFile = FsTreeGet(m.ThumbnailCacheId)
		for page := range m.PageCount {
			m.fullresCacheFiles[page] = FsTreeGet(m.FullresCacheIds[page])
		}

		// Check cache files exist
		if m.thumbCacheFile != nil && m.fullresCacheFiles[0] != nil {
			cacheExists = true
		}

		task.SwLap("Check cache exists")
	}

	m.AddFile(f)
	m.Owner = f.Owner()
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
		if m.MediaType.MultiPage {
			err = m.generateImages()
		} else {
			err = m.generateImage()
		}
		if err != nil {
			util.DisplayError(err)
			return
		}
		task.SwLap("Generate Image")
		err = m.handleCacheCreation(f)
		if err != nil {
			return
		}
		task.SwLap("Create cache")
	}

	if m.RecognitionTags == nil && m.MediaType.SupportsImgRecog {
		err = m.getImageRecognitionTags()
		if err != nil {
			util.DisplayError(err)
		}
		task.SwLap("Get img recognition tags")
	}

	return m, nil
}

func (m *Media) Id() string {
	if m.MediaId == "" {
		if len(m.FileIds) == 0 {
			err := errors.New("trying to generate mediaId for media with no FileId")
			util.DisplayError(err)
			return ""
		}
		err := m.generateFileHash()
		util.DisplayError(err)
	}

	return m.MediaId
}

func (m *Media) IsFilledOut() (bool, string) {
	if m.MediaId == "" {
		return false, "filehash"
	}
	if len(m.FileIds) == 0 {
		return false, "file id"
	}
	if m.Owner == "" {
		return false, "owner"
	}
	if m.MediaType.SupportsImgRecog && m.RecognitionTags == nil {
		return false, "recognition tags"
	}

	// Visual media specific properties
	if m.MediaType != nil && m.MediaType.IsDisplayable {
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
		if m.ThumbLength == 0 {
			return false, "thumb length"
		}
	}

	if m.CreateDate.IsZero() {
		return false, "create date"
	}

	return true, ""
}

func (m *Media) ReadDisplayable(q Quality, index ...int) (data []byte, err error) {
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
		util.DisplayError(err)
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

func (m *Media) AddFile(f *WeblensFile) {
	m.FileIds = util.AddToSet(m.FileIds, []string{f.Id()})
	fddb.addFileToMedia(m, f)
}

func (m *Media) RemoveFile(fId string) {
	var existed bool
	m.FileIds, _, existed = util.YoinkFunc(m.FileIds, func(f string) bool { return f == fId })

	if !existed {
		util.Warning.Println("Attempted to remove file from media that did not have that file")
	}

	if len(m.FileIds) == 0 {
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

func (m *Media) SetImported() {
	if !m.imported {
		m.imported = true
		mediaMapAdd(m)
	}
}

func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}
	return m.imported
}

func (m *Media) WriteToDb() error {
	if !m.imported {
		return fddb.AddMedia(m)
	} else {
		return fddb.UpdateMedia(m)
	}
}

func (m *Media) GetProminentColors() (prom []string, err error) {
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

func (m *Media) loadExif(f *WeblensFile) error {
	if gexift == nil {
		err := errors.New("exiftool not initialized")
		return err
	}
	fileInfos := gexift.ExtractMetadata(f.String())
	if fileInfos[0].Err != nil {
		return fileInfos[0].Err
	}

	m.rawExif = fileInfos[0].Fields
	return nil
}

func (m *Media) parseExif(f *WeblensFile) error {

	if m.CreateDate.Unix() == 0 && m.MediaType != nil && m.imgBytes != nil {
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

	if m.MediaType == nil {
		mimeType, ok := m.rawExif["MIMEType"].(string)
		if !ok {
			mimeType = "generic"
		}
		m.MediaType = ParseMimeType(mimeType)
	}

	if m.MediaType.FriendlyName == "PDF" {
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

	if !m.MediaType.IsDisplayable || m.MediaType.RawThumbExifKey == "" {
		return nil
	}

	raw64 := m.rawExif[m.MediaType.RawThumbExifKey].(string)
	raw64 = raw64[strings.Index(raw64, ":")+1:]

	imgBytes, err := base64.StdEncoding.DecodeString(raw64)
	if err != nil {
		return err
	}
	m.imgBytes = imgBytes

	return nil
}

func (m *Media) readFileBytes(f *WeblensFile) (err error) {
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
	util.DisplayError(err)
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

func (m *Media) generateImages() (err error) {
	if len(m.imgBytes) == 0 {
		return errors.New("cannot generate media image with no imgBytes")
	}
	if m.PageCount == 0 {
		return errors.New("cannot load multipage image without page count")
	}

	m.images = make([]*bimg.Image, m.PageCount)
	var bi *bimg.Image
	for page := range m.PageCount {
		bi = bimg.NewImage(m.imgBytes)
		_, err := bi.Process(bimg.Options{Type: bimg.WEBP, PageNum: page})
		if err != nil {
			return err
		}
		m.images[page] = bi
	}

	imgSize, err := bi.Size()
	if err != nil {
		return
	}

	m.MediaHeight = imgSize.Height
	m.MediaWidth = imgSize.Width

	return nil
}

func (m *Media) generateFileHash() (err error) {
	if m.imgBytes == nil {
		return errors.New("cannot generate media fileHash with no imgBytes")
	}

	h := sha256.New()

	_, err = io.Copy(h, bytes.NewReader(m.imgBytes))
	if err != nil {
		util.DisplayError(err)
		return
	}

	m.MediaId = base64.URLEncoding.EncodeToString(h.Sum(nil))[:8]
	return
}

func (m *Media) generateBlurhash() (err error) {
	return
	// if m.BlurHash == "" {
	// 	m.BlurHash, err = blurhash.Encode(3, 3, m.image)
	// }
	return
}

func (m *Media) getCacheFile(q Quality, generateIfMissing bool, pageNum int) (f *WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	}

	if pageNum >= m.PageCount {
		return nil, ErrPageOutOfRange
	}

	var cacheFileId string
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

func (m *Media) handleCacheCreation(f *WeblensFile) (err error) {
	if len(m.imgBytes) == 0 {
		if m.MediaType.IsRaw {
			err = m.parseExif(f)
			if err != nil {
				util.Debug.Println("Returning with err:", err)
				return
			}
		} else {
			err = m.readFileBytes(f)
			if err != nil {
				util.Debug.Println("Returning with err:", err)
				return
			}
		}
	}

	if m.image == nil {
		err = m.generateImage()
		if err != nil {
			util.Debug.Println("Returning with err:", err)
			return
		}
	}

	thumbW := int((THUMBNAIL_HEIGHT / float32(m.MediaHeight)) * float32(m.MediaWidth))

	var thumbBytes []byte
	if m.image != nil {
		thumbBytes, err = m.image.Resize(thumbW, int(THUMBNAIL_HEIGHT))
		if err != nil {
			util.Debug.Println("Returning with err:", err)
			return
		}
	}

	if len(m.images) != 0 && m.images[0] != nil {
		thumbBytes, err = m.images[0].Resize(thumbW, int(THUMBNAIL_HEIGHT))
		if err != nil {
			util.Debug.Println("Returning with err:", err)
			return
		}
	}

	m.ThumbLength = len(thumbBytes)
	m.cacheDisplayable(Thumbnail, thumbBytes, 0)

	if m.PageCount != 1 {
		for page := range m.PageCount {
			m.cacheDisplayable(Fullres, m.images[page].Image(), page)
		}
	} else {
		m.FullresLength = len(m.imgBytes)
		m.cacheDisplayable(Fullres, m.imgBytes, 0)
	}

	util.Debug.Println("Returning with err:", err)
	return
}

func (m *Media) cacheDisplayable(q Quality, data []byte, pageNum int) *WeblensFile {
	cacheFileName := m.getCacheFilename(q, pageNum)

	f, err := Touch(GetCacheDir(), cacheFileName, true)
	if err != nil && err != ErrFileAlreadyExists {
		util.DisplayError(err)
		return nil
	} else if err == ErrFileAlreadyExists {
		return f
	}

	err = f.Write(data)
	if err != nil {
		util.DisplayError(err)
		return f
	}

	if q == Fullres {
		if m.fullresCacheFiles == nil {
			m.fullresCacheFiles = make([]*WeblensFile, m.PageCount)
		}
		if m.FullresCacheIds == nil {
			m.FullresCacheIds = make([]string, m.PageCount)
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

func (m *Media) getCacheId(q Quality, pageNum int) string {
	if q == Thumbnail && m.ThumbnailCacheId != "" {
		return m.ThumbnailCacheId
	} else if q == Fullres && m.FullresCacheIds[pageNum] != "" {
		return m.FullresCacheIds[pageNum]
	}
	absPath := filepath.Join(GetCacheDir().absolutePath, m.getCacheFilename(q, pageNum))
	return util.GlobbyHash(8, GuaranteeRelativePath(absPath))
}

func (m *Media) getCacheFilename(q Quality, pageNum int) string {
	var cacheFileName string

	if m.PageCount == 1 || q == Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.wlcache", m.Id(), q)
	} else if q != Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.wlcache", m.Id(), q, pageNum)
	}

	return cacheFileName
}

func (m *Media) getImageRecognitionTags() (err error) {
	var imgBuf *bytes.Buffer
	if m.imgBytes == nil {
		var f *WeblensFile
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
	m.RecognitionTags = regocTags

	return
}
