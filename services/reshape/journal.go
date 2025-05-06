package reshape

import (
	"time"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/structs"
)

func FileActionToFileActionInfo(fa history.FileAction) structs.FileActionInfo {
	return structs.FileActionInfo{
		ActionType:      fa.ActionType,
		DestinationPath: fa.DestinationPath.ToPortable(),
		EventId:         fa.EventId,
		FileId:          fa.FileId,
		Filepath:        fa.Filepath.ToPortable(),
		OriginPath:      fa.OriginPath.ToPortable(),
		Size:            fa.Size,
		Timestamp:       fa.Timestamp.UnixMilli(),
		TowerId:         fa.TowerId,
		ContentId:       fa.ContentId,
	}
}

func FileActionInfoToFileAction(info openapi.FileActionInfo) history.FileAction {
	filepath, _ := fs.ParsePortable(info.GetFilepath())
	originPath, _ := fs.ParsePortable(info.GetOriginPath())
	destinationPath, _ := fs.ParsePortable(info.GetDestinationPath())

	return history.FileAction{
		ActionType:      info.ActionType,
		DestinationPath: destinationPath,
		EventId:         info.EventId,
		FileId:          info.FileId,
		Filepath:        filepath,
		OriginPath:      originPath,
		Size:            info.Size,
		Timestamp:       time.UnixMilli(info.Timestamp),
		TowerId:         info.TowerId,
		ContentId:       info.GetContentId(),
	}
}
