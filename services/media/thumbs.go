package media

import (
	"github.com/davidbyttow/govips/v2/vips"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog"
)

func HandleCacheCreation(ctx context_service.AppContext, m *media_model.Media, file *file_model.WeblensFileImpl) (thumbBytes []byte, err error) {
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
			fullHeight = HighresMaxSize * m.Height / m.Width
		} else {
			fullHeight = HighresMaxSize
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
