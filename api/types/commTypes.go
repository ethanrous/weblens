package types

type BroadcasterAgent interface {
	PushFileCreate(newFile WeblensFile)
	PushFileUpdate(updatedFile WeblensFile)
	PushFileMove(preMoveFile WeblensFile, postMoveFile WeblensFile)
	PushFileDelete(deletedFile WeblensFile)
	PushTaskUpdate(taskId TaskId, event TaskEvent, result TaskResult)
	PushShareUpdate(username Username, newShareInfo Share)
	Enable()
	IsBuffered() bool

	FolderSubToTask(folder FileId, task TaskId)
	UnsubTask(task Task)
}

type TaskBroadcaster interface {
	PushTaskUpdate(taskId TaskId, status string, result TaskResult)
}

type BufferedBroadcasterAgent interface {
	BroadcasterAgent
	DropBuffer()
	DisableAutoFlush()
	AutoFlushEnable()
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
