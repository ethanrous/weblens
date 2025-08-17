package media

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
)

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

func NewMediaFromFile(ctx context_service.AppContext, f *file_model.WeblensFileImpl) (m *media_model.Media, err error) {
	if f.GetContentId() == "" {
		return nil, errors.WithStack(file_model.ErrNoContentId)
	}

	m, err = newMedia(ctx, f)
	if err != nil {
		return nil, err
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

	return m, nil
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
