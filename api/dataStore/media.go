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
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/h2non/bimg"
	"github.com/kolesa-team/go-webp/webp"
)

func (m *Media) LoadFromFile(f *WeblensFile) (media *Media, err error) {
	sw := util.NewStopwatch("Load Media")

	err = m.parseExif(f)
	if err != nil {
		return
	}
	sw.Lap("Parse Exif")

	if m.imgBytes == nil {
		err = m.readFileBytes(f)
		if err != nil {
			return
		}
		sw.Lap("Read File")
	}

	if m.MediaId == "" {
		err = m.generateFileHash()
		if err != nil {
			return
		}
		sw.Lap("Generate Filehash")
	}

	storedM, err := MediaMapGet(m.Id())
	if err != nil && err != ErrNoMedia {
		return
	}
	sw.Lap("Check if exists")

	var cacheExists bool
	if storedM != nil {
		err = f.SetMedia(storedM)
		if err != nil {
			return
		}
		sw.Lap("Set file")

		storedM.imgBytes = m.imgBytes
		m = storedM

		m.thumbCacheFile = FsTreeGet(m.ThumbnailCacheId)
		m.fullresCacheFile = FsTreeGet(m.FullresCacheId)

		// Check cache files exist
		if m.thumbCacheFile != nil && m.fullresCacheFile != nil {
			cacheExists = true
			// f.ClearMedia()
		}
		// util.Debug.Println("Media already exists, but cache files are missing, continuing...")

		sw.Lap("Check cache exists")
	}

	m.AddFile(f)
	m.Owner = f.Owner()
	sw.Lap("Add file and set owner")

	// if m.BlurHash == "" || !cacheExists {
	if !cacheExists {
		err = m.generateImage()
		if err != nil {
			return
		}
		sw.Lap("Generate Image")
	}

	// if m.BlurHash == "" {
	// 	err = m.generateBlurhash()
	// 	if err != nil {
	// 		return
	// 	}
	// 	sw.Lap("Generate blurhash")
	// }

	if !cacheExists {
		err = m.handleCacheCreation(f)
		if err != nil {
			return
		}
		sw.Lap("Create cache")
	}

	if m.RecognitionTags == nil && m.MediaType.SupportsImgRecog {
		err = m.getImageRecognitionTags()
		if err != nil {
			util.DisplayError(err)
		}
		sw.Lap("Get img recognition tags")
	}

	sw.Stop()
	sw.PrintResults()

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

func (m *Media) ReadDisplayable(q quality) (data []byte, err error) {
	defer m.Clean()

	var redisKey string
	if util.ShouldUseRedis() {
		redisKey = fmt.Sprintf("%s-%s", m.Id(), q)
		var redisData string
		redisData, err = fddb.RedisCacheGet(redisKey)
		if err == nil {
			data = []byte(redisData)
			return
		}
	}

	// util.Debug.Println("Reading file cached", q)

	f, err := m.getCacheFile(q, true)
	if f == nil || err != nil {
		if err == nil {
			err = errors.New("failed to get cache file")
		}
		return
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
	// m.thumb = nil
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

	var bi *bimg.Image
	if m.MediaType.FriendlyName == "HEIC" {
		return ErrUnsupportedImgType
		// Currently unable to find a cross platform way to support encoding HEIC files
		// Libraries either borked on MacOS (where I dev), or on linux, the target of this
		// app

		// r := bytes.NewReader(m.imgBytes)
		// i, err = goheif.Decode(r)
		// if err != nil {
		// 	return err
		// }
	} else {
		bi = bimg.NewImage(m.imgBytes)
		// i, err = imaging.Decode(bytes.NewReader(m.imgBytes))
		// if err != nil {
		// 	return err
		// }
	}

	// Rotation is backwards because imaging rotates CW, but exif stores CW rotation
	switch m.rotate {
	case "Rotate 270 CW":
		bi.Rotate(90)
		// i = imaging.Rotate90(i)
	case "Rotate 90 CW":
		bi.Rotate(270)
		// i = imaging.Rotate270(i)
	}
	m.image = bi

	b, err := imageToWebp(bi)
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

func (m *Media) calculateThumbSize(thumb image.Image) {
	dimentions := thumb.Bounds()
	width, height := dimentions.Dx(), dimentions.Dy()
	m.MediaHeight = height
	m.MediaWidth = width

	aspectRatio := float64(width) / float64(height)

	var newWidth, newHeight float64

	var bindSize = 800.0
	if aspectRatio > 1 {
		newWidth = bindSize
		newHeight = math.Floor(bindSize / aspectRatio)
	} else {
		newWidth = math.Floor(bindSize * aspectRatio)
		newHeight = bindSize
	}

	if newWidth == 0 || newHeight == 0 {
		panic(fmt.Errorf("thumbnail width or height is 0"))
	}
	m.ThumbWidth = int(newWidth)
	m.ThumbHeight = int(newHeight)

}

func (m *Media) generateBlurhash() (err error) {
	return
	// if m.BlurHash == "" {
	// 	m.BlurHash, err = blurhash.Encode(3, 3, m.image)
	// }
	return
}

// func (m *Media) generateThumbnail() (err error) {
// 	if m.image == nil {
// 		err = errors.New("cannot generate thumbnail with no m.image")
// 		return
// 	}

// 	m.calculateThumbSize(m.image)
// 	thumb := imaging.Thumbnail(m.image, m.ThumbWidth, m.ThumbHeight, imaging.Lanczos)
// 	m.thumb = thumb

// 	return
// }

func (m *Media) getCacheFile(q quality, generateIfMissing bool) (f *WeblensFile, err error) {
	if q == Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	}
	if q == Fullres && m.fullresCacheFile != nil && m.fullresCacheFile.Exists() {
		f = m.fullresCacheFile
		return
	}

	cacheFileId := m.getCacheId(q)
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
		m.ThumbnailCacheId = f.Id()
		m.thumbCacheFile = f
	} else if q == Fullres {
		m.FullresCacheId = f.Id()
		m.fullresCacheFile = f
	}

	return
}

const THUMBNAIL_HEIGHT float32 = 500

func (m *Media) handleCacheCreation(f *WeblensFile) (err error) {
	sw := util.NewStopwatch("Create cache")
	if len(m.imgBytes) == 0 {
		if m.MediaType.IsRaw {
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
	sw.Lap("Got img bytes")

	if m.image == nil {
		err = m.generateImage()
		if err != nil {
			return
		}
	}
	sw.Lap("Got img")

	thumbW := int((THUMBNAIL_HEIGHT / float32(m.MediaHeight)) * float32(m.MediaWidth))
	thumbBytes, err := m.image.Resize(thumbW, int(THUMBNAIL_HEIGHT))
	if err != nil {
		return
	}

	// if m.thumb == nil {
	// 	err = m.generateThumbnail()
	// 	if err != nil {
	// 		return
	// 	}
	// }
	sw.Lap("Got thumbnail")

	// thumbBytes, err := imageToWebp(m.thumb)
	// if err != nil {
	// 	return
	// }
	// sw.Lap("Got thumb bytes")

	// We MUST call generate image before this runs, either preemptively
	// or in the check a few lines above. That will ensure that m.imgBytes are
	// webp encoded and of the fullres image
	fullresBytes := m.imgBytes
	sw.Lap("Got fullres bytes")

	m.ThumbLength = len(thumbBytes)
	m.cacheDisplayable(Thumbnail, thumbBytes)
	sw.Lap("Cached thumb")

	m.FullresLength = len(fullresBytes)
	m.cacheDisplayable(Fullres, fullresBytes)
	sw.Lap("Cached fullres")

	sw.Stop()
	// sw.PrintResults()

	return
}

func (m *Media) cacheDisplayable(q quality, data []byte) *WeblensFile {
	f, err := Touch(GetCacheDir(), fmt.Sprintf("%s-%s.wlcache", m.Id(), q), true)
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

	if q == Thumbnail {
		m.ThumbnailCacheId = f.Id()
		m.thumbCacheFile = f
	} else if q == Fullres {
		m.FullresCacheId = f.Id()
		m.fullresCacheFile = f
	}

	return f
}

func imageToWebp(bi *bimg.Image) (data []byte, err error) {
	data, err = bi.Convert(bimg.WEBP)
	return
}

func (m *Media) getCacheId(q quality) string {
	if q == Thumbnail && m.ThumbnailCacheId != "" {
		return m.ThumbnailCacheId
	} else if q == Fullres && m.FullresCacheId != "" {
		return m.FullresCacheId
	}
	absPath := filepath.Join(GetCacheDir().absolutePath, fmt.Sprintf("%s-%s.wlcache", m.Id(), q))
	return util.GlobbyHash(8, GuaranteeRelativePath(absPath))
}

func (m *Media) getImageRecognitionTags() (err error) {
	var imgBuf *bytes.Buffer
	if m.imgBytes == nil {
		var f *WeblensFile
		f, err = m.getCacheFile(Thumbnail, false)
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
