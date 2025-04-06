package reshape

import (
	"time"

	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/structs"
)

func FileActionToFileActionInfo(fa *history.FileAction) structs.FileActionInfo {
	return structs.FileActionInfo{
		Filepath:        fa.Filepath.ToPortable(),
		OriginPath:      fa.OriginPath.ToPortable(),
		DestinationPath: fa.DestinationPath.ToPortable(),
		Timestamp:       fa.Timestamp.UnixMilli(),
		ActionType:      fa.ActionType,
		EventId:         fa.EventId,
		Size:            fa.Size,
		TowerId:         fa.TowerId,
	}
}

func FileActionInfoToFileAction(info structs.FileActionInfo) *history.FileAction {
	filepath, _ := fs.ParsePortable(info.OriginPath)
	originPath, _ := fs.ParsePortable(info.OriginPath)
	destinationPath, _ := fs.ParsePortable(info.DestinationPath)

	return &history.FileAction{
		Timestamp:       time.UnixMilli(info.Timestamp),
		ActionType:      info.ActionType,
		Filepath:        filepath,
		OriginPath:      originPath,
		DestinationPath: destinationPath,
		EventId:         info.EventId,
		Size:            info.Size,
		TowerId:         info.TowerId,
	}
}
