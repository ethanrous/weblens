package reshape

import (
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/structs"
)

func FileActionToFileActionInfo(fa *history.FileAction) structs.FileActionInfo {
	return structs.FileActionInfo{
		Timestamp:       fa.Timestamp.UnixMilli(),
		ActionType:      fa.ActionType,
		OriginPath:      fa.OriginPath,
		DestinationPath: fa.DestinationPath,
		LifeId:          fa.LifeId,
		EventId:         fa.EventId,
		Size:            fa.Size,
		ParentId:        fa.ParentId,
		ServerId:        fa.ServerId,
	}
}
