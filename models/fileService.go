package models

import (
	"io"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/task"
)

type FileService interface {
	GetUserFile(id fileTree.FileId) (*fileTree.WeblensFileImpl, error)
	GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, error)
	GetFileSafe(id fileTree.FileId, accessor *User, share *FileShare) (*fileTree.WeblensFileImpl, error)
	GetMediaRoot() *fileTree.WeblensFileImpl
	PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error)

	CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent) (*fileTree.WeblensFileImpl, error)
	CreateFolder(parent *fileTree.WeblensFileImpl, folderName string, caster FileCaster) (
		*fileTree.WeblensFileImpl, error,
	)
	ImportFile(f *fileTree.WeblensFileImpl) error

	GetFileOwner(file *fileTree.WeblensFileImpl) *User
	IsFileInTrash(file *fileTree.WeblensFileImpl) bool

	MoveFiles(files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, caster FileCaster) error
	RenameFile(file *fileTree.WeblensFileImpl, newName string, caster FileCaster) error
	MoveFilesToTrash(file []*fileTree.WeblensFileImpl, mover *User, share *FileShare, caster FileCaster) error
	ReturnFilesFromTrash(files []*fileTree.WeblensFileImpl, caster FileCaster) error
	DeleteFiles(files []*fileTree.WeblensFileImpl, caster FileCaster) error
	RestoreFiles(ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster FileCaster) error

	ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error)

	GetMediaCacheByFilename(filename string) (*fileTree.WeblensFileImpl, error)
	NewCacheFile(media *Media, quality MediaQuality, pageNum int) (*fileTree.WeblensFileImpl, error)
	DeleteCacheFile(file fileTree.WeblensFile) error

	AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error
	RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error
	GetTasks(f *fileTree.WeblensFileImpl) []*task.Task

	GetUsersJournal() fileTree.Journal

	ResizeDown(file *fileTree.WeblensFileImpl, caster FileCaster) error
	ResizeUp(file *fileTree.WeblensFileImpl, caster FileCaster) error
	NewZip(zipName string, owner *User) (*fileTree.WeblensFileImpl, error)
	GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error)
}

type FileCaster interface {
	PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media)
	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
	PushFileCreate(newFile *fileTree.WeblensFileImpl)
	PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl)

	PushFileDelete(deletedFile *fileTree.WeblensFileImpl)
	Close()
}
