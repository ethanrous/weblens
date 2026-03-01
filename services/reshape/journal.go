package reshape

import (
	"time"

	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlstructs"
)

// FileActionToFileActionInfo converts a FileAction to a FileActionInfo structure suitable for API responses.
func FileActionToFileActionInfo(fa history.FileAction) wlstructs.FileActionInfo {
	return wlstructs.FileActionInfo{
		ActionType:      fa.ActionType,
		DestinationPath: fa.DestinationPath.ToPortable(),
		EventID:         fa.EventID,
		FileID:          fa.FileID,
		Filepath:        fa.Filepath.ToPortable(),
		OriginPath:      fa.OriginPath.ToPortable(),
		Size:            fa.Size,
		Timestamp:       fa.Timestamp.UnixMilli(),
		TowerID:         fa.TowerID,
		ContentID:       fa.ContentID,
	}
}

// FileActionInfoToFileAction converts a FileActionInfo from the API to a FileAction.
func FileActionInfoToFileAction(info wlstructs.FileActionInfo) history.FileAction {
	filepath, _ := wlfs.ParsePortable(info.Filepath)
	originPath, _ := wlfs.ParsePortable(info.OriginPath)
	destinationPath, _ := wlfs.ParsePortable(info.DestinationPath)

	return history.FileAction{
		ActionType:      info.ActionType,
		DestinationPath: destinationPath,
		EventID:         info.EventID,
		FileID:          info.FileID,
		Filepath:        filepath,
		OriginPath:      originPath,
		Size:            info.Size,
		Timestamp:       time.UnixMilli(info.Timestamp),
		TowerID:         info.TowerID,
		ContentID:       info.ContentID,
	}
}
