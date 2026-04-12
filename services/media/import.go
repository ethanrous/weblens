package media

import (
	"time"

	"github.com/ethanrous/agno/bindings/go/agno"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

func newMedia(ctx context_service.AppContext, f *file_model.WeblensFileImpl) (*media_model.Media, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, f)
	if err != nil {
		return nil, err
	}

	return &media_model.Media{
		ContentID: f.GetContentID(),
		Owner:     ownerName,
		FileIDs:   []string{f.ID()},
		LikedBy:   []string{},
		Enabled:   true,
	}, nil
}

// NewMediaFromFile creates a new Media object from a file by extracting metadata from EXIF data.
func NewMediaFromFile(ctx context_service.AppContext, f *file_model.WeblensFileImpl) (m *media_model.Media, err error) {
	img, err := agno.Open(f.GetPortablePath().ToAbsolute())
	if err != nil {
		return nil, err
	}

	if f.GetContentID() == "" {
		return nil, wlerrors.WithStack(file_model.ErrNoContentID)
	}

	m, err = newMedia(ctx, f)
	if err != nil {
		return nil, err
	}

	if m.CreateDate.Unix() <= 0 {
		createDate, err := getCreateDateFromExif(img, f)
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
			width, err := agno.ExifValue[int](img, agno.ImageWidth)
			if err != nil {
				return nil, err
			}

			m.Width = width

			height, err := agno.ExifValue[int](img, agno.ImageHeight)
			if err != nil {
				return nil, err
			}

			m.Height = height

			duration, err := getVideoDurationMs(f.GetPortablePath().ToAbsolute())
			if err != nil {
				return nil, err
			}

			m.Duration = duration
		}
	}

	if m.Location[0] == 0 && m.Location[1] == 0 {
		loc, err := img.GPSCoordinates()
		if err != nil {
			ctx.Log().Warn().Msgf("failed to get GPS coordinates from EXIF for file %s: %v", f.ID(), err)
		} else {
			m.Location = loc
		}
	}

	mType := GetMediaType(m)
	if !mType.IsSupported() {
		return nil, media_model.ErrMediaBadMimeType
	}

	m.PageCount = img.PageCount()

	return m, nil
}

func loadImageFromFile(f *file_model.WeblensFileImpl, _ media_model.MType) (*agno.Image, error) {
	filePath := f.GetPortablePath().ToAbsolute()

	img, err := agno.Open(filePath)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func getCreateDateFromExif(img *agno.Image, file *file_model.WeblensFileImpl) (createDate time.Time, err error) {
	r, err := agno.ExifValue[string](img, agno.CreateDate)
	if err != nil {
		r, err = agno.ExifValue[string](img, agno.DateTimeOriginal)
	}

	if err != nil {
		r, err = agno.ExifValue[string](img, agno.ModifyDate)
	}

	if err != nil {
		wlog.GlobalLogger().Warn().Msgf("failed to get date from EXIF for file %s: %v", file.ID(), err)

		return file.ModTime(), nil
	}

	offset, _ := agno.ExifValue[string](img, agno.OffsetTime)

	dateFormats := []string{
		"2006:01:02 15:04:05.000-07:00",
		"2006:01:02 15:04:05.00-07:00",
		"2006:01:02 15:04:05-07:00",
		// Some EXIF data may not include timezone information, so we try parsing without it as well, and then apply the offset time if available
		"2006:01:02 15:04:05",
	}

	for _, format := range dateFormats {
		createDate, err = time.Parse(format, r)
		if err == nil {
			return createDate, nil
		}

		if offset != "" {
			createDate, err = time.Parse(format, r+offset)
			if err == nil {
				return createDate, nil
			}
		}
	}

	return file.ModTime(), nil
}
