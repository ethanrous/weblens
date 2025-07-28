package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/davidbyttow/govips/v2/vips"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetConverted(ctx context.Context, m *media_model.Media, format media_model.MediaType) ([]byte, error) {
	if !format.IsMime("image/jpeg") {
		return nil, errors.New("unsupported format")
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, errors.WithStack(context_service.ErrNoContext)
	}

	file, err := appCtx.FileService.GetFileById(ctx, m.FileIDs[0])
	if err != nil {
		return nil, err
	}

	img, err := loadImageFromFile(file, format)
	if err != nil {
		return nil, err
	}

	bs, _, err := img.ExportJpeg(&vips.JpegExportParams{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	log.FromContext(ctx).Debug().Msgf("Exported %s to jpeg", m.ID())

	return bs, nil
}

var exifd *exiftool.Exiftool

type cacheKey string

const (
	CacheIdKey      cacheKey = "cacheId"
	CacheQualityKey cacheKey = "cacheQuality"
	CachePageKey    cacheKey = "cachePageNum"
	CacheMediaKey   cacheKey = "cacheMedia"

	HighresMaxSize = 2500
	ThumbMaxSize   = 500

	exifToolByfferSize = 100_000
)

var extraMimes = []struct{ ext, mime string }{
	{ext: ".m3u8", mime: "application/vnd.apple.mpegurl"},
	{ext: ".mp4", mime: "video/mp4"},
}

func mediaServiceStartup(context.Context, config.ConfigProvider) error {
	for _, em := range extraMimes {
		err := mime.AddExtensionType(em.ext, em.mime)

		if err != nil {
			return err
		}
	}

	var err error
	exifd, err = exiftool.NewExiftool(
		exiftool.Api("largefilesupport"),
		// 					    	100 KB
		exiftool.Buffer([]byte{}, exifToolByfferSize),
	)

	if err != nil {
		panic(err)
	}

	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(&vips.Config{})

	return nil
}

func init() {
	startup.RegisterStartup(mediaServiceStartup)
}

func FetchCacheImg(ctx context_service.AppContext, m *media_model.Media, q media_model.MediaQuality, pageNum int) ([]byte, error) {
	cacheKey := m.ContentID + string(q) + strconv.Itoa(pageNum)
	cache := ctx.GetCache("photoCache")

	anyBs, ok := cache.Get(cacheKey)
	if ok {
		return anyBs.([]byte), nil
	}

	f, err := getCacheFile(ctx, m, q, pageNum)

	if err != nil {
		return nil, err
	}

	bs, err := f.ReadAll()
	if err != nil {
		return nil, err
	}

	cache.Set(cacheKey, bs)

	return bs, nil
}

var ErrMediaNotVideo = errors.New("media is not a video")

const videoStreamerContextKey = "videoStreamerContextKey"

func StreamVideo(ctx context.Context, m *media_model.Media) (*media_model.VideoStreamer, error) {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, errors.WithStack(context_service.ErrNoContext)
	}

	if !media_model.ParseMime(m.MimeType).IsVideo {
		return nil, errors.WithStack(ErrMediaNotVideo)
	}

	cache := appCtx.GetCache(videoStreamerContextKey)

	streamerAny, ok := cache.Get(m.ID())
	if ok {
		return streamerAny.(*media_model.VideoStreamer), nil
	}

	f, err := appCtx.FileService.GetFileById(ctx, m.FileIDs[0])
	if err != nil {
		return nil, err
	}

	streamer := media_model.NewVideoStreamer(f, file_model.ThumbsDirPath)

	cache.Set(m.ID(), streamer)

	return streamer, nil
}

//	func (ms *MediaServiceImpl) SetMediaLiked(mediaId models.ContentId, liked bool, username string) error {
//		m := ms.Get(mediaId)
//		if m == nil {
//			return errors.Errorf("Could not find media with id [%s] while trying to update liked array", mediaId)
//		}
//
//		filter := bson.M{"contentId": mediaId}
//		var update bson.M
//		if liked && len(m.LikedBy) == 0 {
//			update = bson.M{"$set": bson.M{"likedBy": []string{username}}}
//		} else if liked && len(m.LikedBy) == 0 {
//			update = bson.M{"$addToSet": bson.M{"likedBy": username}}
//		} else {
//			update = bson.M{"$pull": bson.M{"likedBy": username}}
//		}
//
//		_, err := ms.collection.UpdateOne(context.Background(), filter, update)
//		if err != nil {
//			return err
//		}
//
//		if liked {
//			m.LikedBy = wl_slices.AddToSet(m.LikedBy, username)
//		} else {
//			m.LikedBy = wl_slices.Filter(
//				m.LikedBy, func(u string) bool {
//					return u != username
//				},
//			)
//		}
//
//		return nil
//	}
//
//	func (ms *MediaServiceImpl) GetMediaConverted(m *models.Media, format string) ([]byte, error) {
//		f, err := ms.fileService.GetFileByTree(m.FileIDs[0], UsersTreeKey)
//		if err != nil {
//			return nil, err
//		}
//
//		img, err := ms.loadImageFromFile(f, ms.GetMediaType(m))
//		if err != nil {
//			return nil, err
//		}
//
//		var blob []byte
//		switch format {
//		case "png":
//			blob, _, err = img.ExportPng(nil)
//		case "jpeg":
//			blob, _, err = img.ExportJpeg(nil)
//		default:
//			return nil, errors.Errorf("Unknown media convert format [%s]", format)
//		}
//		return blob, err
//	}
//
//	func (ms *MediaServiceImpl) removeCacheFiles(media *models.Media) error {
//		thumbCache, err := ms.getCacheFile(media, models.LowRes, 0)
//		if err != nil && !errors.Is(err, errors.ErrNoFile) {
//			return err
//		}
//
//		if thumbCache != nil {
//			err = ms.fileService.DeleteCacheFile(thumbCache)
//			if err != nil {
//				return err
//			}
//		}
//
//		highresCacheFile, err := ms.getCacheFile(media, models.HighRes, 0)
//		if err != nil && !errors.Is(err, errors.ErrNoFile) {
//			return err
//		}
//
//		if highresCacheFile != nil {
//			err = ms.fileService.DeleteCacheFile(highresCacheFile)
//			if err != nil {
//				return err
//			}
//		}
//
//		return nil
//	}

func newMedia(ctx context_service.AppContext, f *file_model.WeblensFileImpl) (*media_model.Media, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, f)
	if err != nil {
		return nil, err
	}

	return &media_model.Media{
		ContentID:       f.GetContentId(),
		Owner:           ownerName,
		FileIDs:         []string{f.ID()},
		RecognitionTags: []string{},
		LikedBy:         []string{},
		Enabled:         true,
	}, nil
}

// Regex to capture degrees, minutes, seconds, and direction
var coordRe = regexp.MustCompile(`(\d+)\s*deg\s*(\d+)'\s*([\d\.]+)"\s*([NSEW])`)

// Converts a coordinate string to decimal degrees
func parseCoordinate(coord string) (float64, error) {
	matches := coordRe.FindStringSubmatch(coord)
	if len(matches) != 5 {
		return 0, fmt.Errorf("invalid coordinate format")
	}

	degrees, _ := strconv.ParseFloat(matches[1], 64)
	minutes, _ := strconv.ParseFloat(matches[2], 64)
	seconds, _ := strconv.ParseFloat(matches[3], 64)
	dir := matches[4]

	decimal := degrees + minutes/60 + seconds/3600
	if dir == "S" || dir == "W" {
		decimal = -decimal
	}

	return decimal, nil
}

// Parses a full coordinate string
func getDecimalCoords(coordStr string) (lat, lon float64, err error) {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("coordinate string must have two parts separated by a comma")
	}

	lat, err = parseCoordinate(strings.TrimSpace(parts[0]))
	if err != nil {
		return
	}

	lon, err = parseCoordinate(strings.TrimSpace(parts[1]))

	return
}

func NewMediaFromFile(ctx context_service.AppContext, f *file_model.WeblensFileImpl) (m *media_model.Media, err error) {
	if f.GetContentId() == "" {
		return nil, errors.WithStack(file_model.ErrNoContentId)
	}

	m, err = media_model.GetMediaByContentId(ctx, f.GetContentId())
	if err != nil {
		m, err = newMedia(ctx, f)
		if err != nil {
			return nil, err
		}
	}

	fileMetas := exifd.ExtractMetadata(f.GetPortablePath().ToAbsolute())

	for _, fileMeta := range fileMetas {
		if fileMeta.Err != nil {
			return nil, fileMeta.Err
		}
	}

	if m.CreateDate.Unix() <= 0 {
		createDate, err := getCreateDateFromExif(fileMetas[0].Fields, f)
		if err != nil {
			return nil, err
		}

		m.CreateDate = createDate
	}

	if m.MimeType == "" {
		ext := f.GetPortablePath().Ext()
		mType := media_model.ParseExtension(ext)
		m.MimeType = mType.Mime

		if media_model.ParseMime(m.MimeType).IsVideo {
			m.Width = int(fileMetas[0].Fields["ImageWidth"].(float64))
			m.Height = int(fileMetas[0].Fields["ImageHeight"].(float64))

			duration, err := getVideoDurationMs(f.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}

			m.Duration = duration
		}
	}

	if m.Location[0] == 0 || m.Location[1] == 0 {
		pos, ok := fileMetas[0].Fields["GPSPosition"].(string)
		if ok {
			lat, long, err := getDecimalCoords(pos)
			if err != nil {
				return nil, err
			}

			m.Location[0] = lat
			m.Location[1] = long
		}
	}

	mType := GetMediaType(m)
	if !mType.IsSupported() {
		return nil, media_model.ErrMediaBadMimeType
	}

	if mType.IsMultiPage() {
		m.PageCount = int(fileMetas[0].Fields["PageCount"].(float64))
	} else {
		m.PageCount = 1
	}

	_, err = handleCacheCreation(ctx, m, f)
	if err != nil {
		return m, err
	}

	_, err = getHighDimensionImageEncoding(ctx, m)
	if err != nil {
		return m, err
	}

	return m, nil
}

func GetMediaType(m *media_model.Media) media_model.MediaType {
	return media_model.ParseMime(m.MimeType)
}

func handleCacheCreation(ctx context_service.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) (thumbBytes []byte, err error) {
	mType := GetMediaType(m)

	if !mType.IsVideo {
		img, err := loadImageFromFile(file, mType)
		if err != nil {
			return nil, err
		}

		m.PageCount = img.Pages()
		// Read image dimensions
		m.Height = img.Height()
		m.Width = img.Width()

		if mType.IsMultiPage() {
			fullPdf, err := file.ReadAll()
			if err != nil {
				return nil, errors.WithStack(err)
			}

			for page := range m.PageCount {
				vipsPage := vips.IntParameter{}
				vipsPage.Set(page)

				img, err := vips.LoadImageFromBuffer(fullPdf, &vips.ImportParams{Page: vipsPage})
				if err != nil {
					return nil, errors.WithStack(err)
				}

				err = handleNewHighRes(ctx, m, img, page)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err = handleNewHighRes(ctx, m, img, 0)
			if err != nil {
				return nil, err
			}
		}

		// Resize thumb image if too big
		if m.Width > ThumbMaxSize || m.Height > ThumbMaxSize {
			var thumbWidth, thumbHeight uint
			if m.Width > m.Height {
				thumbWidth = ThumbMaxSize
				thumbHeight = uint(float64(ThumbMaxSize) / float64(m.Width) * float64(m.Height))
			} else {
				thumbHeight = ThumbMaxSize
				thumbWidth = uint(float64(ThumbMaxSize) / float64(m.Height) * float64(m.Width))
			}

			ctx.Log().Trace().Func(func(e *zerolog.Event) {
				e.Msgf("Resizing %s thumb image to %dx%d", file.GetPortablePath(), thumbWidth, thumbHeight)
			})

			err = img.Resize(float64(thumbHeight)/float64(m.Height), vips.KernelAuto)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		// Create and write thumb cache file
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, errors.WithStack(err)
		} else if err == nil {
			blob, _, err := img.ExportWebp(nil)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			_, err = thumb.Write(blob)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			m.SetLowresCacheFile(thumb)

			thumbBytes = blob
		}
	} else {
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, errors.WithStack(err)
		} else if err == nil {
			thumbBytes, err = generateVideoThumbnail(file.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}

			_, err = thumb.Write(thumbBytes)
			if err != nil {
				return nil, err
			}

			m.SetLowresCacheFile(thumb)
		}
	}

	return thumbBytes, nil
}

func handleNewHighRes(ctx context_service.AppContext, m *media_model.Media, img *vips.ImageRef, page int) error {
	// Resize highres image if too big
	if m.Width > HighresMaxSize || m.Height > HighresMaxSize {
		var fullHeight int
		if m.Width > m.Height {
			// fullWidth = HighresMaxSize
			fullHeight = HighresMaxSize * m.Height / m.Width
		} else {
			fullHeight = HighresMaxSize
			// fullWidth = HighresMaxSize * m.Width / m.Height
		}

		err := img.Resize(float64(fullHeight)/float64(m.Height), vips.KernelAuto)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// Create and write highres cache file
	highres, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.HighRes), page)
	if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
		return err
	} else if err == nil {
		params := &vips.WebpExportParams{NearLossless: true, Quality: 100}

		blob, _, err := img.ExportWebp(params)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = highres.Write(blob)
		if err != nil {
			return err
		}

		m.SetHighresCacheFiles(highres, page)
	}

	return nil
}

func getCacheFile(ctx context_service.AppContext, m *media_model.Media, quality media_model.MediaQuality, pageNum int) (*file_model.WeblensFileImpl, error) {
	// if quality == media_model.LowRes && m.GetLowresCacheFile() != nil {
	// 	return m.GetLowresCacheFile(), nil
	// } else if quality == media_model.HighRes && m.GetHighresCacheFiles(pageNum) != nil {
	// 	return m.GetHighresCacheFiles(pageNum), nil
	// }

	filename := fmtCacheFileName(m, quality, pageNum)

	cacheFile, err := ctx.FileService.GetMediaCacheByFilename(ctx, filename)
	if err != nil {
		return nil, err
	}

	// if quality == media_model.LowRes {
	// 	m.SetLowresCacheFile(cacheFile)
	// } else if quality == media_model.HighRes {
	// 	m.SetHighresCacheFiles(cacheFile, pageNum)
	// } else {
	// 	return nil, errors.Errorf("Unknown media quality [%s]", quality)
	// }

	return cacheFile, nil
}

func fmtCacheFileName(m *media_model.Media, quality media_model.MediaQuality, pageNum int) string {
	var pageNumStr string
	if pageNum > 1 && quality == media_model.HighRes {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}

	filename := fmt.Sprintf("%s-%s%s.cache", m.ID(), quality, pageNumStr)

	return filename
}

func loadImageFromFile(f *file_model.WeblensFileImpl, mType media_model.MediaType) (*vips.ImageRef, error) {
	filePath := f.GetPortablePath().ToAbsolute()

	var img *vips.ImageRef

	var err error

	// Sony RAWs do not play nice with govips. Should fall back to imagick but it thinks its a TIFF.
	// The real libvips figures this out, adding an intermediary step using dcraw to convert to a real TIFF
	// and continuing processing from there solves this issue, and is surprisingly fast. Everyone say "Thank you dcraw"
	if strings.HasSuffix(filePath, "ARW") || strings.HasSuffix(filePath, "CR2") {
		cmd := exec.Command("dcraw", "-T", "-w", "-h", "-c", filePath)

		var stdb, errb bytes.Buffer
		cmd.Stderr = &errb
		cmd.Stdout = &stdb

		err = cmd.Run()
		if err != nil {
			return nil, errors.WithStack(errors.New(err.Error() + "\n" + errb.String()))
		}

		img, err = vips.NewImageFromReader(&stdb)
	} else {
		img, err = vips.NewImageFromFile(filePath)
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	// PDFs and HEIFs do not need to be rotated.
	if !mType.IsMultiPage() && !mType.IsMime("image/heif") {
		// Rotate image based on exif data
		err = img.AutoRotate()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return img, nil
}

func getCreateDateFromExif(exif map[string]any, file *file_model.WeblensFileImpl) (createDate time.Time, err error) {
	r, ok := exif["SubSecCreateDate"]
	if !ok {
		r, ok = exif["MediaCreateDate"]
	}

	if ok {
		createDate, err = time.Parse("2006:01:02 15:04:05.000-07:00", r.(string))
		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05.00-07:00", r.(string))
		}

		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05", r.(string))
		}

		if err != nil {
			createDate, err = time.Parse("2006:01:02 15:04:05-07:00", r.(string))
		}

		if err != nil {
			createDate = file.ModTime()
		}
	} else {
		createDate = file.ModTime()
	}

	return createDate, nil
}

func generateVideoThumbnail(filepath string) ([]byte, error) {
	const frameNum = 10

	buf := bytes.NewBuffer(nil)
	errOut := bytes.NewBuffer(nil)

	// Get the 10th frame of the video and save it to the cache as the thumbnail
	// "Highres" for video is the video itself
	err := ffmpeg.Input(filepath).Filter(
		"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
	).Output(
		"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
	).WithOutput(buf).WithErrorOutput(errOut).Run()
	if err != nil {
		return nil, errors.WithStack(errors.New(err.Error() + errOut.String()))
	}

	return buf.Bytes(), nil
}

func getVideoDurationMs(filepath string) (int, error) {
	probeJson, err := ffmpeg.Probe(filepath)
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
		return 0, errors.Errorf("invalid movie format")
	}

	duration, err := strconv.ParseFloat(formatChunk["duration"].(string), 32)
	if err != nil {
		return 0, err
	}

	return int(duration) * 1000, nil
}

type MediaWithScore struct {
	Media *media_model.Media
	Score float64
}

// Returns the highest values above the largest significant gap in the array.
// If no significant gap, returns all values.
func skimTop(values []MediaWithScore) []MediaWithScore {
	valueRange := values[0].Score - values[len(values)-1].Score
	minScore := values[len(values)-1].Score + valueRange*0.50 // Only scores in the top 50%
	log.GlobalLogger().Debug().Msgf("Skimming top values with min score: %f %f", values[0].Score, minScore)

	lowestItemIndex := slices.IndexFunc(values, func(ms MediaWithScore) bool {
		return ms.Score < minScore
	})

	values = values[:lowestItemIndex]

	return values
}

func SortMediaByTextSimilarity(ctx context_service.AppContext, search string, ms []*media_model.Media, minScore float64) ([]MediaWithScore, error) {
	if len(search) == 0 {
		return []MediaWithScore{}, nil
	}

	scores, err := getSimilarityScores(ctx, search, ms...)
	if err != nil {
		return nil, err
	}

	if len(scores) != len(ms) {
		return nil, errors.Errorf("expected %d similarity scores, got %d", len(ms), len(scores))
	}

	msScores := slices.MapI(ms, func(m *media_model.Media, i int) MediaWithScore {
		return MediaWithScore{
			m,
			scores[i],
		}
	})

	msScores = slices.Filter(msScores, func(ms MediaWithScore) bool {
		return ms.Score >= minScore
	})

	slices.SortFunc(msScores, func(a, b MediaWithScore) int {
		if a.Score < b.Score {
			return 1
		} else if a.Score > b.Score {
			return -1
		}

		return 0
	})

	msScores = skimTop(msScores)

	ctx.Log().Debug().Msgf("Similarity scores: %v", msScores)

	return msScores, nil
}

const hdieServerUrl = "http://weblens-hdir:5000"

func getHighDimensionImageEncoding(ctx context_service.AppContext, m *media_model.Media) ([]float64, error) {
	f, err := getCacheFile(ctx, m, media_model.LowRes, 0)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(hdieServerUrl + "/encode?img-path=" + f.GetPortablePath().String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	target := []float64{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return nil, err
	}

	m.HDIR = target

	return target, nil
}

func getSimilarityScores(ctx context_service.AppContext, text string, m ...*media_model.Media) ([]float64, error) {
	hdirs := [][]float64{}
	for _, media := range m {
		if len(media.HDIR) == 0 {
			hdir, err := getHighDimensionImageEncoding(ctx, media)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			hdirs = append(hdirs, hdir)
		} else {
			hdirs = append(hdirs, media.HDIR)
		}
	}

	hdirBytes, err := json.Marshal(hdirs)
	if err != nil {
		return nil, err
	}

	reqBody := bytes.NewBuffer((fmt.Appendf(nil, `{"text": "%s", "image_features": %s}`, text, hdirBytes)))

	resp, err := http.Post(hdieServerUrl+"/match", "application/json", reqBody)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	target := struct {
		Similarity []float64 `json:"similarity"`
	}{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		ctx.Log().Debug().Msgf("Error unmarshalling similarity response: %s", reqBody)

		return nil, err
	}

	return target.Similarity, nil
}
