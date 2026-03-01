package media

import (
	"context"
	"mime"
	"strconv"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/media/agno"
)

type cacheKey string

// ErrMediaNotVideo indicates that the media is not a video.
var ErrMediaNotVideo = wlerrors.New("media is not a video")

// Cache keys
const (
	CacheIDKey      cacheKey = "cacheID"
	CacheQualityKey cacheKey = "cacheQuality"
	CachePageKey    cacheKey = "cachePageNum"
	CacheMediaKey   cacheKey = "cacheMedia"

	videoStreamerContextKey = "videoStreamerContextKey"

	HighresMaxSize = 2500
	ThumbMaxSize   = 500
)

var extraMimes = []struct{ ext, mime string }{
	{ext: ".m3u8", mime: "application/vnd.apple.mpegurl"},
	{ext: ".mp4", mime: "video/mp4"},
}

func init() {
	startup.RegisterHook(mediaServiceStartup)
}

func mediaServiceStartup(_ context.Context, _ config.Provider) error {
	for _, em := range extraMimes {
		err := mime.AddExtensionType(em.ext, em.mime)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetConverted returns the media in the specified format, currently only supports JPEG.
// The quality parameter controls JPEG encoding quality (1-100).
func GetConverted(ctx context.Context, m *media_model.Media, format media_model.MType, quality int) ([]byte, error) {
	if !format.IsMime("image/jpeg") {
		return nil, wlerrors.New("unsupported format")
	}

	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, wlerrors.WithStack(context_service.ErrNoContext)
	}

	file, err := appCtx.FileService.GetFileByID(ctx, m.FileIDs[0])
	if err != nil {
		return nil, err
	}

	img, err := loadImageFromFile(file, format)
	if err != nil {
		return nil, err
	}
	defer img.Free()

	bs, err := agno.WriteJpeg(img, quality)
	if err != nil {
		return nil, err
	}

	wlog.FromContext(ctx).Debug().Msgf("Exported %s to jpeg (quality=%d)", m.ID(), quality)

	return bs, nil
}

// IsCached checks if the media is fully cached, meaning low-res and all high-res thumbs are available.
func IsCached(ctx context_service.AppContext, m *media_model.Media) (bool, error) {
	lowres, err := getCacheFile(ctx, m, media_model.LowRes, 0)
	if wlerrors.Is(err, file_model.ErrFileNotFound) {
		return false, nil
	} else if err != nil {
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
func FetchCacheImg(ctx context_service.AppContext, m *media_model.Media, q media_model.Quality, pageNum int) ([]byte, error) {
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

// GetMediaType returns the media type by parsing the media's MIME type.
func GetMediaType(m *media_model.Media) media_model.MType {
	return media_model.ParseMime(m.MimeType)
}
