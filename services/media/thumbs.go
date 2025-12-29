package media

import (
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/media/agno"
)

// HandleCacheCreation creates media cache files (thumbs and highres) for the given media and file.
func HandleCacheCreation(ctx context_service.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) (thumbBytes []byte, err error) {
	mType := GetMediaType(m)

	if !mType.IsVideo {
		img, err := loadImageFromFile(file, mType)
		if err != nil {
			return nil, err
		}

		defer img.Free()

		m.PageCount = 1

		// Read image dimensions
		m.Width, m.Height = img.Dimensions()
		ctx.Log().Debug().Msgf("Loaded image dimensions for %s: %dx%d (pages: %d)", file.GetPortablePath(), m.Width, m.Height, m.PageCount)

		if mType.IsMultiPage() {
			return nil, errors.New("multi-page media not yet supported")
		}

		err = handleNewHighRes(ctx, m, img, 0)
		if err != nil {
			return nil, err
		}

		// Resize thumb image if too big
		if m.Width > ThumbMaxSize || m.Height > ThumbMaxSize {
			var thumbHeight uint
			if m.Width > m.Height {
				thumbHeight = uint(float64(ThumbMaxSize) / float64(m.Width) * float64(m.Height))
			} else {
				thumbHeight = ThumbMaxSize
			}

			err = img.Resize(float64(thumbHeight) / float64(m.Height))
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		// Create and write thumb cache file
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, errors.WithStack(err)
		} else if err == nil {
			err = agno.WriteWebp(thumb.GetPortablePath().ToAbsolute(), img)
			if err != nil {
				return nil, err
			}

			m.SetLowresCacheFile(thumb)
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

func handleNewHighRes(ctx context_service.AppContext, m *media_model.Media, img *agno.Image, page int) error {
	// Resize highres image if too big
	if m.Width > HighresMaxSize || m.Height > HighresMaxSize {
		var fullHeight int
		if m.Width > m.Height {
			fullHeight = int(float64(HighresMaxSize) * float64(m.Height) / float64(m.Width))
		} else {
			fullHeight = HighresMaxSize
		}

		err := img.Resize(float64(fullHeight) / float64(m.Height))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// Create and write highres cache file
	highres, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.HighRes), page)
	if err != nil && !errors.Is(err, file_model.ErrFileAlreadyExists) {
		return err
	} else if err == nil {
		err = agno.WriteWebp(highres.GetPortablePath().ToAbsolute(), img)
		if err != nil {
			return err
		}

		m.SetHighresCacheFiles(highres, page)
	}

	return nil
}

func getCacheFile(ctx context_service.AppContext, m *media_model.Media, quality media_model.Quality, pageNum int) (*file_model.WeblensFileImpl, error) {
	filename, err := media_model.FmtCacheFileName(m.ID(), quality, pageNum)
	if err != nil {
		return nil, err
	}

	cacheFile, err := ctx.FileService.GetMediaCacheByFilename(ctx, filename)
	if err != nil {
		return nil, err
	}

	return cacheFile, nil
}
