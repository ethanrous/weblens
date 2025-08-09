package media

import (
	"context"
	"mime"
	"strconv"

	"github.com/barasher/go-exiftool"
	"github.com/davidbyttow/govips/v2/vips"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/startup"
	context_service "github.com/ethanrous/weblens/services/context"
)

var exifd *exiftool.Exiftool

type cacheKey string

var ErrMediaNotVideo = errors.New("media is not a video")

const (
	CacheIdKey      cacheKey = "cacheId"
	CacheQualityKey cacheKey = "cacheQuality"
	CachePageKey    cacheKey = "cachePageNum"
	CacheMediaKey   cacheKey = "cacheMedia"

	videoStreamerContextKey = "videoStreamerContextKey"

	HighresMaxSize = 2500
	ThumbMaxSize   = 500

	exifToolByfferSize = 100_000 // 100 KB
)

var extraMimes = []struct{ ext, mime string }{
	{ext: ".m3u8", mime: "application/vnd.apple.mpegurl"},
	{ext: ".mp4", mime: "video/mp4"},
}

func init() {
	startup.RegisterStartup(mediaServiceStartup)
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
		exiftool.Buffer([]byte{}, exifToolByfferSize),
	)

	if err != nil {
		panic(err)
	}

	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(&vips.Config{})

	return nil
}

// GetConverted returns the media in the specified format, currently only supports JPEG.
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

// IsCached checks if the media is fully cached, meaning low-res and all high-res thumbs are available.
func IsCached(ctx context_service.AppContext, m *media_model.Media) (bool, error) {
	lowres, err := getCacheFile(ctx, m, media_model.LowRes, 0)
	if err != nil {
		return false, err
	}

	if lowres == nil {
		return false, nil
	}

	for page := range m.PageCount {
		highres, err := getCacheFile(ctx, m, media_model.HighRes, page)
		if err != nil {
			return false, err
		}

		if highres == nil {
			return false, nil
		}
	}

	return true, nil
}

// FetchCacheImg retrieves the cached image for the given media, quality, and page number.
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

func GetMediaType(m *media_model.Media) media_model.MediaType {
	return media_model.ParseMime(m.MimeType)
}
