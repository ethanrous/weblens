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
	Flush()
}

type Requester interface {
	GetCoreSnapshot() error
}
