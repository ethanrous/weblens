package models

import (
	"io"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/task"
)

type FileService interface {
	GetFile(id fileTree.FileId) (*fileTree.WeblensFile, error)
	GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFile, error)
	GetFileSafe(id fileTree.FileId, accessor *User, share *FileShare) (*fileTree.WeblensFile, error)

	GetFileOwner(file *fileTree.WeblensFile) *User
	IsFileInTrash(file *fileTree.WeblensFile) bool

	ImportFile(f *fileTree.WeblensFile) error

	MoveFiles(files []*fileTree.WeblensFile, destFolder *fileTree.WeblensFile, caster FileCaster,) error
	RenameFile(file *fileTree.WeblensFile, newName string, caster FileCaster) error
	MoveFileToTrash(file *fileTree.WeblensFile, mover *User, share *FileShare, caster FileCaster) error
	ReturnFilesFromTrash(files []*fileTree.WeblensFile, caster FileCaster) error
	PermanentlyDeleteFiles(files []*fileTree.WeblensFile, caster FileCaster) error

	ReadFile(f *fileTree.WeblensFile) (io.ReadCloser, error)

	GetThumbFileName(filename string) (*fileTree.WeblensFile, error)
	NewCacheFile(contentId string, quality MediaQuality, pageNum int) (*fileTree.WeblensFile, error)
	DeleteCacheFile(file *fileTree.WeblensFile) error

	AddTask(f *fileTree.WeblensFile, t *task.Task) error
	RemoveTask(f *fileTree.WeblensFile, t *task.Task) error
	GetTasks(f *fileTree.WeblensFile) []*task.Task

	GetMediaJournal() fileTree.JournalService

	ResizeDown(file *fileTree.WeblensFile, caster FileCaster) error
	ResizeUp(file *fileTree.WeblensFile, caster FileCaster) error
	NewZip(zipName string, owner *User) (*fileTree.WeblensFile, error)
}

type FileCaster interface {
	PushFileUpdate(updatedFile *fileTree.WeblensFile, media *Media)
	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
	PushFileCreate(newFile *fileTree.WeblensFile)
	PushFileMove(preMoveFile *fileTree.WeblensFile, postMoveFile *fileTree.WeblensFile)

	PushFileDelete(deletedFile *fileTree.WeblensFile)
	Close()
}
