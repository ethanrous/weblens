package media

import (
	"github.com/ethanrous/agno/bindings/go/agno"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// HandleCacheCreation creates media cache files (thumbs and highres) for the given media and file.
func HandleCacheCreation(ctx context_service.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) (thumbBytes []byte, err error) {
	mType := GetMediaType(m)

	if !mType.IsVideo {
		if mType.IsMultiPage() {
			return handleMultiPageCache(ctx, m, file)
		}

		img, err := loadImageFromFile(file, mType)
		if err != nil {
			return nil, err
		}

		defer img.Close() //nolint:errcheck

		m.PageCount = 1

		// Read image dimensions
		m.Width, m.Height = img.Dimensions()

		img, err = handleNewHighRes(ctx, m, img, 0)
		if err != nil {
			return nil, err
		}

		img, err = resizeToFit(img, ThumbMaxSize)
		if err != nil {
			return nil, wlerrors.WithStack(err)
		}

		// Create and write thumb cache file
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !wlerrors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, wlerrors.WithStack(err)
		} else if err == nil {
			err = img.WriteWebP(thumb.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}

			m.SetLowresCacheFile(thumb)
		}
	} else {
		thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
		if err != nil && !wlerrors.Is(err, file_model.ErrFileAlreadyExists) {
			return nil, wlerrors.WithStack(err)
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

// handleMultiPageCache creates cache files for all pages of a multi-page document (e.g., PDF).
// High-res WebP files are created for every page. The thumbnail is generated from page 0.
func handleMultiPageCache(ctx context_service.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) ([]byte, error) {
	filePath := file.GetPortablePath().ToAbsolute()

	img, err := agno.OpenPage(filePath, 0, 0, 0)
	if err != nil {
		return nil, wlerrors.Errorf("failed to load page 0 of %s: %w", file.GetPortablePath(), err)
	}

	defer func() {
		if img != nil {
			img.Close() //nolint:errcheck
		}
	}()

	m.Width, m.Height = img.Dimensions()
	m.PageCount = img.PageCount()

	ctx.Log().Debug().Msgf(
		"Loaded multi-page media %s: %dx%d (%d pages)",
		file.GetPortablePath(), m.Width, m.Height, m.PageCount,
	)

	img, err = handleNewHighRes(ctx, m, img, 0)
	if err != nil {
		return nil, err
	}

	img, err = resizeToFit(img, ThumbMaxSize)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	thumb, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.LowRes), 0)
	if err != nil && !wlerrors.Is(err, file_model.ErrFileAlreadyExists) {
		return nil, wlerrors.WithStack(err)
	} else if err == nil {
		err = img.WriteWebP(thumb.GetPortablePath().ToAbsolute())
		if err != nil {
			return nil, err
		}

		m.SetLowresCacheFile(thumb)
	}

	img.Close() //nolint:errcheck
	img = nil

	for page := 1; page < m.PageCount; page++ {
		pageImg, err := agno.OpenPage(filePath, page, 0, 0)
		if err != nil {
			return nil, wlerrors.Errorf("failed to load page %d of %s: %w", page, file.GetPortablePath(), err)
		}

		resultImg, err := handleNewHighRes(ctx, m, pageImg, page)
		if resultImg != nil {
			resultImg.Close() //nolint:errcheck
		}

		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func handleNewHighRes(ctx context_service.AppContext, m *media_model.Media, img *agno.Image, page int) (*agno.Image, error) {
	var err error

	img, err = resizeToFit(img, HighresMaxSize)
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	highres, err := ctx.FileService.NewCacheFile(m.ID(), string(media_model.HighRes), page)
	if err != nil && !wlerrors.Is(err, file_model.ErrFileAlreadyExists) {
		return img, err
	} else if err == nil {
		err = img.WriteWebP(highres.GetPortablePath().ToAbsolute())
		if err != nil {
			return img, err
		}

		m.SetHighresCacheFiles(highres, page)
	}

	return img, nil
}

// resizeToFit scales img so its largest dimension fits within maxSize, preserving aspect ratio.
// Returns the original img unchanged if it already fits. The receiver is consumed if resizing occurs.
func resizeToFit(img *agno.Image, maxSize int) (*agno.Image, error) {
	w, h := img.Dimensions()
	if w <= maxSize && h <= maxSize {
		return img, nil
	}

	var targetH int
	if w > h {
		targetH = int(float64(maxSize) * float64(h) / float64(w))
	} else {
		targetH = maxSize
	}

	return img.Resize(float64(targetH) / float64(h))
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
