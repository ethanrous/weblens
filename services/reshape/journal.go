package reshape

func FileActionToFileActionInfo(fa *fileTree.FileAction) FileActionInfo {
	return FileActionInfo{
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
