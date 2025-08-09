package reshape

import (
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/structs"
)

func MediaToMediaInfo(m *media_model.Media) structs.MediaInfo {
	return structs.MediaInfo{
		MediaId:         m.MediaID.Hex(),
		ContentId:       m.ContentID,
		FileIds:         m.FileIDs,
		CreateDate:      m.CreateDate.UnixMilli(),
		Owner:           m.Owner,
		Width:           m.Width,
		Height:          m.Height,
		PageCount:       m.PageCount,
		Duration:        m.Duration,
		MimeType:        m.MimeType,
		Location:        m.Location,
		RecognitionTags: m.GetRecognitionTags(),
		Hidden:          m.Hidden,
		Enabled:         m.Enabled,
		LikedBy:         m.LikedBy,
		Imported:        m.IsImported(),
	}
}

func NewMediaBatchInfo(m []*media_model.Media) structs.MediaBatchInfo {
	if len(m) == 0 {
		return structs.MediaBatchInfo{
			Media:      []structs.MediaInfo{},
			MediaCount: 0,
		}
	}

	mediaInfos := make([]structs.MediaInfo, 0, len(m))
	for _, media := range m {
		mediaInfos = append(mediaInfos, MediaToMediaInfo(media))
	}

	return structs.MediaBatchInfo{
		Media:      mediaInfos,
		MediaCount: len(m),
	}
}

func MediaTypeToMediaTypeInfo(mt media_model.MediaType) structs.MediaTypeInfo {
	return structs.MediaTypeInfo{
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
