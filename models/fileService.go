package models

import (
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/task"
)

type FileService interface {
	GetFileSafe(id fileTree.FileId, accessor *User, share *FileShare) (*fileTree.WeblensFile, error)
	GetFileOwner(file *fileTree.WeblensFile) *User

	IsFileInTrash(file *fileTree.WeblensFile) bool
	MoveFileToTrash(file *fileTree.WeblensFile, mover *User, share *FileShare, caster FileCaster) error
	ReturnFilesFromTrash(files []*fileTree.WeblensFile, caster FileCaster) error
	MoveFile(
		file *fileTree.WeblensFile, destFolder *fileTree.WeblensFile, newFilename string,
		caster FileCaster,
	) error
	GetMediaJournal() fileTree.JournalService
	AttachFile(file *fileTree.WeblensFile, caster FileCaster) error

	ResizeDown(file *fileTree.WeblensFile, caster FileCaster) error
	ResizeUp(file *fileTree.WeblensFile, caster FileCaster) error
	NewZip(zipName string, owner *User) (*fileTree.WeblensFile, error)
}

type FileCaster interface {
	// PushWeblensEvent(eventTag string)
	PushFileUpdate(updatedFile *fileTree.WeblensFile)
	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
	PushFileCreate(newFile *fileTree.WeblensFile)
	PushFileMove(preMoveFile *fileTree.WeblensFile, postMoveFile *fileTree.WeblensFile)

	PushFileDelete(deletedFile *fileTree.WeblensFile)
	// PushShareUpdate(username Username, newShareInfo Share)
	// Enable()
	// Disable()
	// IsEnabled() bool
	// IsBuffered() bool
	//
	// FolderSubToTask(folder fileTree.FileId, taskId task.TaskId)
	// FolderSubToPool(folder fileTree.FileId, poolId task.TaskId)
	// UnsubTask(task *task.Task)
	// DisableAutoFlush()
	// AutoFlushEnable()
	// Flush()
	//
	// Relay(msg WsResponseInfo)
	//
	// // Close flush, release the auto-flusher, and disable the caster
	Close()
}
