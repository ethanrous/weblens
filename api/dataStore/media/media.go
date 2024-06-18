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
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/kolesa-team/go-webp/webp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	MediaId          primitive.ObjectID `json:"-" bson:"_id"`
	ContentId        types.ContentId    `json:"mediaId" bson:"mediaId"`
	FileIds          []types.FileId     `json:"fileIds" bson:"fileIds"`
	FullresCacheIds  []types.FileId     `json:"fullresCacheIds" bson:"fullresCacheIds"`
	ThumbnailCacheId types.FileId       `json:"thumbnailCacheId" bson:"thumbnailCacheId"`
	CreateDate       time.Time          `json:"createDate" bson:"createDate"`
	Owner            types.User         `json:"owner" bson:"owner"`
	MediaWidth       int                `json:"mediaWidth" bson:"mediaWidth"`
	MediaHeight      int                `json:"mediaHeight" bson:"mediaHeight"`
	PageCount        int                `json:"pageCount" bson:"pageCount"`
	BlurHash         string             `json:"blurHash" bson:"blurHash"`
	MimeType         string             `json:"mimeType" bson:"mimeType"`
	RecognitionTags  []string           `json:"recognitionTags" bson:"recognitionTags"`
	Hidden           bool               `json:"hidden" bson:"hidden"`
	Enabled          bool               `json:"enabled" bson:"enabled"`

	mediaType types.MediaType
	imported  bool

	rotate string
	image  *bimg.Image
	images []*bimg.Image

	rawExif           map[string]any
	thumbCacheFile    types.WeblensFile
	fullresCacheFiles []types.WeblensFile
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

	err = m.handleCacheCreation(f)
	if err != nil {
		return
	}
	task.SwLap("Create cache")

	if m.RecognitionTags == nil && m.mediaType.SupportsImgRecog() {
		err = m.getImageRecognitionTags(f.GetTree())
		if err != nil {
			util.ErrTrace(err)
		}
		task.SwLap("Get img recognition tags")
	}

	m.AddFile(f)

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

func (m *Media) SetOwner(owner types.User) {
	m.Owner = owner
}

func (m *Media) GetOwner() types.User {
	return m.Owner
}

func (m *Media) ReadDisplayable(q types.Quality, ft types.FileTree, index ...int) (data []byte, err error) {
	var pageNum int
	if len(index) != 0 && (index[0] != 0 && index[0] >= m.PageCount) {
		return nil, dataStore.ErrPageOutOfRange
	} else if len(index) != 0 {
		pageNum = index[0]
	} else {
		pageNum = 0
	}

	return types.SERV.MediaRepo.FetchCacheImg(m, q, pageNum, ft)
}

func (m *Media) GetFiles() []types.FileId {
	return m.FileIds
}

func (m *Media) AddFile(f types.WeblensFile) error {
	m.FileIds = util.AddToSet(m.FileIds, []types.FileId{f.ID()})
	return types.SERV.Database.AddFileToMedia(m.ID(), f.ID())
}

func (m *Media) RemoveFile(f types.WeblensFile) error {
	var existed bool
	m.FileIds, _, existed = util.YoinkFunc(m.FileIds, func(existFile types.FileId) bool { return existFile == f.ID() })

	if !existed {
		util.Warning.Println("Attempted to remove file from Media that did not have that file")
	}

	err := types.SERV.Database.RemoveFileFromMedia(m.ID(), f.ID())
	if err != nil {
		return err
	}

	if len(m.FileIds) == 0 {
		err := m.Hide()
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
	if !m.imported && i {
		m.imported = true
		err := types.SERV.MediaRepo.Add(m)
		if err != nil {
			util.ShowErr(err)
			return
		}
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

func (m *Media) IsCached(ft types.FileTree) bool {
	if m.thumbCacheFile == nil {
		if m.ThumbnailCacheId == "" {
			return false
		}
		cache := ft.Get(m.ThumbnailCacheId)
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
				cache := ft.Get(m.FullresCacheIds[i])
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

func (m *Media) Hide() error {
	m.Hidden = true
	return types.SERV.Database.HideMedia(m.ID())
}

func (m *Media) getProminentColors(ft types.FileTree) (prom []string, err error) {
	var i image.Image
	thumbBytes, err := m.ReadDisplayable(dataStore.Thumbnail, ft)
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
		m.mediaType = types.SERV.MediaRepo.TypeService().ParseMime(mimeType)
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

	if !m.mediaType.IsDisplayable() || m.mediaType.GetThumbExifKey() == "" {
		return nil
	}

	return nil
}

func (m *Media) generateImage(bs []byte) (err error) {
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

func (m *Media) generateImages(bs []byte) (err error) {
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

func (m *Media) GetCacheFile(q types.Quality, generateIfMissing bool, pageNum int, ft types.FileTree) (f types.WeblensFile, err error) {
	if q == dataStore.Thumbnail && m.thumbCacheFile != nil && m.thumbCacheFile.Exists() {
		f = m.thumbCacheFile
		return
	} else if q == dataStore.Fullres && len(m.fullresCacheFiles) > pageNum && m.fullresCacheFiles[pageNum] != nil {
		f = m.fullresCacheFiles[pageNum]
		return
	}

	if pageNum >= m.PageCount {
		return nil, dataStore.ErrPageOutOfRange
	}

	cacheRoot := ft.Get("CACHE")

	var cacheFileId types.FileId
	if q == dataStore.Fullres && len(m.FullresCacheIds) > pageNum && m.FullresCacheIds[pageNum] != "" {
		cacheFileId = m.FullresCacheIds[pageNum]
	} else if q == dataStore.Fullres {
		cacheFileId = m.getCacheId(q, pageNum, cacheRoot)
		m.FullresCacheIds = make([]types.FileId, m.PageCount)
		m.FullresCacheIds[pageNum] = cacheFileId
	}

	if q == dataStore.Thumbnail {
		if m.ThumbnailCacheId == "" {
			m.ThumbnailCacheId = m.getCacheId(q, pageNum, cacheRoot)
		}
		cacheFileId = m.ThumbnailCacheId
	}

	f = ft.Get(cacheFileId)
	if f == nil || !f.Exists() {
		if generateIfMissing {
			realFile := ft.Get(m.FileIds[0])
			if err != nil {
				return nil, err
			}

			err = m.handleCacheCreation(realFile)
			if err != nil {
				return nil, err
			}

			return m.GetCacheFile(q, false, pageNum, ft)

		} else {
			return nil, dataStore.ErrNoCache
		}
	}

	if q == dataStore.Thumbnail {
		m.thumbCacheFile = f
	} else if q == dataStore.Fullres {
		m.fullresCacheFiles[pageNum] = f
	}

	return
}

const ThumbnailHeight float32 = 500

func (m *Media) handleCacheCreation(f types.WeblensFile) (err error) {
	_, err = m.GetCacheFile(dataStore.Thumbnail, false, 0, f.GetTree())
	hasThumbCache := err == nil
	_, err = m.GetCacheFile(dataStore.Fullres, false, 0, f.GetTree())
	hasFullresCache := err == nil

	if hasThumbCache && hasFullresCache && m.MediaWidth != 0 && m.MediaHeight != 0 {
		return nil
	}

	var bs []byte

	if m.mediaType.IsRaw() {
		util.Debug.Println()
		raw64 := m.rawExif[m.mediaType.GetThumbExifKey()].(string)
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
	}

	if m.mediaType.IsMultiPage() && len(m.images) != 0 && m.images[0] != nil {
		thumbImg := bimg.NewImage(m.images[0].Image())

		thumbBytes, err = thumbImg.Resize(thumbW, int(ThumbnailHeight))
		if err != nil {
			return
		}
	}

	if !hasThumbCache {
		m.cacheDisplayable(dataStore.Thumbnail, thumbBytes, 0, f.GetTree())
	}

	if !hasFullresCache {
		if m.mediaType.IsMultiPage() {
			for page := range m.PageCount {
				m.cacheDisplayable(dataStore.Fullres, m.images[page].Image(), page, f.GetTree())
			}
		} else {
			m.cacheDisplayable(dataStore.Fullres, m.image.Image(), 0, f.GetTree())
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
	if err != nil && !errors.Is(err, types.ErrFileAlreadyExists) {
		util.ErrTrace(err)
		return nil
	} else if errors.Is(err, types.ErrFileAlreadyExists) {
		return f
	}

	err = f.Write(data)
	if err != nil {
		util.ErrTrace(err)
		return f
	}

	if q == dataStore.Thumbnail && m.ThumbnailCacheId == "" || m.thumbCacheFile == nil {
		m.ThumbnailCacheId = f.ID()
		m.thumbCacheFile = f
	} else if q == dataStore.Fullres && m.FullresCacheIds[pageNum] == "" || m.fullresCacheFiles[pageNum] == nil {
		m.FullresCacheIds[pageNum] = f.ID()
		m.fullresCacheFiles[pageNum] = f
	}

	return f
}

func (m *Media) getCacheId(q types.Quality, pageNum int, cacheDir types.WeblensFile) types.FileId {
	return cacheDir.GetTree().GenerateFileId(filepath.Join(cacheDir.GetAbsPath(), m.getCacheFilename(q,
		pageNum)))
}

func (m *Media) getCacheFilename(q types.Quality, pageNum int) string {
	var cacheFileName string

	if m.PageCount == 1 || q == dataStore.Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s.cache", m.ID(), q)
	} else if q != dataStore.Thumbnail {
		cacheFileName = fmt.Sprintf("%s-%s_%d.cache", m.ID(), q, pageNum)
	}

	return cacheFileName
}

func (m *Media) getImageRecognitionTags(ft types.FileTree) (err error) {
	bs, err := m.ReadDisplayable(dataStore.Thumbnail, ft)
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
		"createDate":       m.CreateDate,
		"mimeType":         m.MimeType,
		"recognitionTags":  m.RecognitionTags,
		"pageCount":        m.PageCount,
	}
	return bson.Marshal(data)
}

func (m *Media) UnmarshalBSON(bs []byte) error {
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
		return errors.New("no contentId while decoding Media")
	}

	m.FileIds = util.Map(data["fileIds"].(primitive.A), func(a any) types.FileId { return types.FileId(a.(string)) })
	m.ThumbnailCacheId = types.FileId(data["thumbnailCacheId"].(string))

	if data["fullresCacheIds"] != nil {
		m.FullresCacheIds = util.Map(data["fullresCacheIds"].(primitive.A), func(a any) types.FileId { return types.FileId(a.(string)) })
	}

	m.BlurHash = data["blurHash"].(string)
	m.Owner = types.SERV.UserService.Get(types.Username(data["owner"].(string)))
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
