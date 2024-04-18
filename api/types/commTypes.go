package types

type BroadcasterAgent interface {
	PushFileCreate(newFile WeblensFile)
	PushFileUpdate(updatedFile WeblensFile)
	PushFileMove(preMoveFile WeblensFile, postMoveFile WeblensFile)
	PushFileDelete(deletedFile WeblensFile)
	PushTaskUpdate(taskId TaskId, status string, result TaskResult)
	PushShareUpdate(username Username, newShareInfo Share)
	Enable()
	IsBuffered() bool
}

type TaskBroadcaster interface {
	PushTaskUpdate(taskId TaskId, status string, result TaskResult)
}

type BufferedBroadcasterAgent interface {
	BroadcasterAgent
	DropBuffer()
	DisableAutoflush()
	AutoflushEnable()
	// Flush()

	// flush, release the auto-flusher, and disable the caster
	Close()
}

type Requester interface {
	RequestCoreSnapshot() ([]FileJournalEntry, error)
	AttachToCore(srvId, coreAddress, name string, key WeblensApiKey) error
	GetCoreUsers() (us []User, err error)
	PingCore() bool
	GetCoreFileBin(f WeblensFile) ([][]byte, error)
	GetCoreFileInfos(fIds []FileId) ([]WeblensFile, error)
}
