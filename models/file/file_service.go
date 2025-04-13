package file

import (
	"time"

	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/fs"
)

type FileService interface {
	AddFile(context context.ContextZ, file ...*WeblensFileImpl) error

	// Size returns the size of the specified file tree
	Size(treeAlias string) int64

	// GetFileById retrieves a file by its ID
	GetFileById(fileId string) (*WeblensFileImpl, error)

	// GetFileByFilepath retrieves a file by its filepath
	GetFileByFilepath(ctx context.ContextZ, path fs.Filepath) (*WeblensFileImpl, error)

	// GetFileByContentId retrieves a file by its content ID
	GetFileByContentId(contentId string) (*WeblensFileImpl, error)

	// GetFiles retrieves multiple files by their IDs
	GetFiles(ids []string) ([]*WeblensFileImpl, []string, error)

	// PathToFile resolves a path to a file
	// PathToFile(searchPath string) (*WeblensFileImpl, error)

	// UserPathToFile resolves a user path to a file
	// UserPathToFile(searchPath string, user *user_model.User) (*WeblensFileImpl, error)

	// CreateFile creates a new file
	CreateFile(parent *WeblensFileImpl, filename string, data ...[]byte) (*WeblensFileImpl, error)

	// CreateFolder creates a new folder
	CreateFolder(ctx context.ContextZ, parent *WeblensFileImpl, folderName string) (*WeblensFileImpl, error)

	// CreateUserHome creates a home directory for a user
	CreateUserHome(ctx context.ContextZ, user *user_model.User) error

	// IsFileInTrash checks if a file is in the trash
	// IsFileInTrash(file *WeblensFileImpl) bool

	// MoveFiles moves files to a new location
	MoveFiles(ctx context.ContextZ, files []*WeblensFileImpl, destFolder *WeblensFileImpl, treeName string) error

	// RenameFile renames a file
	RenameFile(file *WeblensFileImpl, newName string) error

	// MoveFilesToTrash moves files to the trash
	MoveFilesToTrash(ctx context.ContextZ, files []*WeblensFileImpl, user *user_model.User, share *share_model.FileShare) error

	// ReturnFilesFromTrash restores files from the trash
	ReturnFilesFromTrash(ctx context.ContextZ, trashFiles []*WeblensFileImpl) error

	// DeleteFiles permanently deletes files
	DeleteFiles(ctx context.ContextZ, files []*WeblensFileImpl) error

	// RestoreFiles restores files from history
	RestoreFiles(ctx context.ContextZ, ids []string, newParent *WeblensFileImpl, restoreTime time.Time) error

	// RestoreHistory restores file history
	// RestoreHistory(lifetimes []*fileTree.Lifetime) error

	// GetMediaCacheByFilename retrieves media cache by filename
	GetMediaCacheByFilename(filename string) (*WeblensFileImpl, error)

	// NewCacheFile creates a new cache file for media
	NewCacheFile(mediaId, quality string, pageNum int) (*WeblensFileImpl, error)

	// DeleteCacheFile deletes a cache file
	DeleteCacheFile(file *WeblensFileImpl) error

	// GetFolderCover gets the cover for a folder
	// GetFolderCover(folder *WeblensFileImpl) (string, error)

	// SetFolderCover sets the cover for a folder
	// SetFolderCover(folderId string, coverId string) error

	// AddTask adds a task to a file
	// AddTask(f *WeblensFileImpl, t *task.Task) error

	// RemoveTask removes a task from a file
	// RemoveTask(f *WeblensFileImpl, t *task.Task) error

	// GetTasks gets tasks associated with a file
	// GetTasks(f *WeblensFileImpl) []*task.Task

	// GetJournalByTree gets the journal for a tree
	// GetJournalByTree(treeName string) fileTree.Journal

	// ResizeDown updates size metadata down the tree
	// ResizeDown(file *WeblensFileImpl, event *fileTree.FileEvent) error

	// ResizeUp updates size metadata up the tree
	// ResizeUp(file *WeblensFileImpl, event *fileTree.FileEvent) error

	// GetThumbsDir gets the thumbnails directory
	// GetThumbsDir() (*WeblensFileImpl, error)

	// NewZip creates a new zip file
	NewZip(zipName string, owner *user_model.User) (*WeblensFileImpl, error)

	// GetZip retrieves a zip file by ID
	GetZip(id string) (*WeblensFileImpl, error)
}
