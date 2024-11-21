package models

import (
	"io"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/task"
)

type FileService interface {
	Size(treeAlias string) int64

	AddTree(tree fileTree.FileTree)

	GetFileByTree(id fileTree.FileId, treeAlias string) (*fileTree.WeblensFileImpl, error)
	GetFileByContentId(contentId ContentId) (*fileTree.WeblensFileImpl, error)
	GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, []fileTree.FileId, error)
	GetFileSafe(id fileTree.FileId, accessor *User, share *FileShare) (*fileTree.WeblensFileImpl, error)

	GetUsersRoot() *fileTree.WeblensFileImpl
	PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error)
	UserPathToFile(searchPath string, user *User) (*fileTree.WeblensFileImpl, error)

	CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent) (*fileTree.WeblensFileImpl, error)
	CreateFolder(parent *fileTree.WeblensFileImpl, folderName string, caster FileCaster) (*fileTree.WeblensFileImpl, error)
	CreateUserHome(user *User) error

	// CreateRestoreFile creates a new file on the restore tree based on the existing user file,
	// it's name will be the contentId of the user file, and it's parent will be the restore tree root.
	// CreateRestoreFile(lifetime *fileTree.Lifetime) (restoreFile *fileTree.WeblensFileImpl, err error)

	NewBackupFile(lt *fileTree.Lifetime) (*fileTree.WeblensFileImpl, error)

	GetFileOwner(file *fileTree.WeblensFileImpl) *User
	IsFileInTrash(file *fileTree.WeblensFileImpl) bool

	MoveFiles(files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, treeName string, caster FileCaster) error
	RenameFile(file *fileTree.WeblensFileImpl, newName string, caster FileCaster) error
	MoveFilesToTrash(file []*fileTree.WeblensFileImpl, mover *User, share *FileShare, caster FileCaster) error
	ReturnFilesFromTrash(files []*fileTree.WeblensFileImpl, caster FileCaster) error
	DeleteFiles(files []*fileTree.WeblensFileImpl, treeName string, caster FileCaster) error
	RestoreFiles(ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster FileCaster) error
	RestoreHistory(lifetimes []*fileTree.Lifetime) error

	ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error)

	GetMediaCacheByFilename(filename string) (*fileTree.WeblensFileImpl, error)
	NewCacheFile(media *Media, quality MediaQuality, pageNum int) (*fileTree.WeblensFileImpl, error)
	DeleteCacheFile(file fileTree.WeblensFile) error

	GetFolderCover(folder *fileTree.WeblensFileImpl) (ContentId, error)
	SetFolderCover(folderId fileTree.FileId, coverId ContentId) error

	AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error
	RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error
	GetTasks(f *fileTree.WeblensFileImpl) []*task.Task

	GetJournalByTree(treeName string) fileTree.Journal

	ResizeDown(file *fileTree.WeblensFileImpl, caster FileCaster) error
	ResizeUp(file *fileTree.WeblensFileImpl, caster FileCaster) error
	NewZip(zipName string, owner *User) (*fileTree.WeblensFileImpl, error)
	GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error)
}

type FileCaster interface {
	PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media)
	PushFilesUpdate(updatedFiles []*fileTree.WeblensFileImpl, medias []*Media)

	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
	PushFileCreate(newFile *fileTree.WeblensFileImpl)

	PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl)
	PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl)

	PushFileDelete(deletedFile *fileTree.WeblensFileImpl)
	PushFilesDelete(deletedFiles []*fileTree.WeblensFileImpl)
	Close()
}
