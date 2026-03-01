package reshape

import (
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/wlstructs"
)

// MediaBatchOptions defines options for creating a batch of media information.
type MediaBatchOptions struct {
	Scores []float64
}

// MediaToMediaInfo converts a Media model to a MediaInfo transfer object.
func MediaToMediaInfo(m *media_model.Media) wlstructs.MediaInfo {
	return wlstructs.MediaInfo{
		MediaID:    m.MediaID.Hex(),
		ContentID:  m.ContentID,
		FileIDs:    m.FileIDs,
		CreateDate: m.CreateDate.UnixMilli(),
		Owner:      m.Owner,
		Width:      m.Width,
		Height:     m.Height,
		PageCount:  m.PageCount,
		Duration:   m.Duration,
		MimeType:   m.MimeType,
		Location:   m.Location,
		Hidden:     m.Hidden,
		Enabled:    m.Enabled,
		LikedBy:    m.LikedBy,
		Imported:   m.IsImported(),
	}
}

// NewMediaBatchInfo creates a batch media information object from a slice of Media models.
func NewMediaBatchInfo(m []*media_model.Media, opts ...MediaBatchOptions) wlstructs.MediaBatchInfo {
	options := MediaBatchOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	if len(m) == 0 {
		return wlstructs.MediaBatchInfo{
			Media:      []wlstructs.MediaInfo{},
			MediaCount: 0,
		}
	}

	mediaInfos := make([]wlstructs.MediaInfo, 0, len(m))
	for i, media := range m {
		info := MediaToMediaInfo(media)
		if i < len(options.Scores) {
			info.HDIRScore = options.Scores[i]
		}

		mediaInfos = append(mediaInfos, info)
	}

	return wlstructs.MediaBatchInfo{
		Media:      mediaInfos,
		MediaCount: len(m),
	}
}

// MediaTypeToMediaTypeInfo converts a MediaType model to a MediaTypeInfo transfer object.
func MediaTypeToMediaTypeInfo(mt media_model.MType) wlstructs.MediaTypeInfo {
	return wlstructs.MediaTypeInfo{
		Mime:            mt.Mime,
		Name:            mt.Name,
		RawThumbExifKey: mt.RawThumbExifKey,
		Extensions:      mt.Extensions,
		Displayable:     mt.Displayable,
		Raw:             mt.Raw,
		Video:           mt.IsVideo,
		ImgRecog:        mt.ImgRecog,
		MultiPage:       mt.MultiPage,
	}
}
