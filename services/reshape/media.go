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
	var mediaInfos []structs.MediaInfo
	for _, media := range m {
		mediaInfos = append(mediaInfos, MediaToMediaInfo(media))
	}
	return structs.MediaBatchInfo{
		Media:      mediaInfos,
		MediaCount: len(m),
	}
}
