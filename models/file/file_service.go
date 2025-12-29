package file

import (
	"context"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/fs"
)

// Service provides operations for managing files and folders in the Weblens system.
// It handles file creation, modification, deletion, and retrieval, as well as media caching and backup operations.
type Service interface {
	AddFile(context context.Context, file ...*WeblensFileImpl) error

	// Size returns the size of the specified file tree
	Size(treeAlias string) int64

	// GetFileByID retrieves a file by its ID
	GetFileByID(ctx context.Context, fileID string) (*WeblensFileImpl, error)

	// GetFileByFilepath retrieves a file by its filepath
	GetFileByFilepath(ctx context.Context, path fs.Filepath, dontLoadNew ...bool) (*WeblensFileImpl, error)

	// CreateFile creates a new file
	CreateFile(ctx context.Context, parent *WeblensFileImpl, filename string, data ...[]byte) (*WeblensFileImpl, error)

	// CreateFolder creates a new folder
	CreateFolder(ctx context.Context, parent *WeblensFileImpl, folderName string) (*WeblensFileImpl, error)

	// GetChildren retrieves children of a folder
	GetChildren(ctx context.Context, folder *WeblensFileImpl) ([]*WeblensFileImpl, error)

	// GetChildrenByPath loads children of a folder recursively
	RecursiveEnsureChildrenLoaded(ctx context.Context, folder *WeblensFileImpl) error

	// CreateUserHome creates a home directory for a user
	CreateUserHome(ctx context.Context, user *user_model.User) error

	// NewBackupRestoreFile creates a new file for backup restoration from a remote tower
	NewBackupRestoreFile(ctx context.Context, contentID, remoteTowerID string) (*WeblensFileImpl, error)

	// InitBackupDirectory initializes the backup directory for a tower
	InitBackupDirectory(ctx context.Context, tower tower_model.Instance) (*WeblensFileImpl, error)
	// IsFileInTrash checks if a file is in the trash
	// IsFileInTrash(file *WeblensFileImpl) bool

	// MoveFiles moves files to a new location
	MoveFiles(ctx context.Context, files []*WeblensFileImpl, destFolder *WeblensFileImpl) error

	// RenameFile renames a file
	RenameFile(ctx context.Context, file *WeblensFileImpl, newName string) error

	// ReturnFilesFromTrash restores files from the trash
	ReturnFilesFromTrash(ctx context.Context, trashFiles []*WeblensFileImpl) error

	// DeleteFiles permanently deletes files
	DeleteFiles(ctx context.Context, files ...*WeblensFileImpl) error

	// RestoreFiles restores files from history
	RestoreFiles(ctx context.Context, ids []string, newParent *WeblensFileImpl, restoreTime time.Time) error

	// RestoreHistory restores file history
	// RestoreHistory(lifetimes []*fileTree.Lifetime) error

	// GetMediaCacheByFilename retrieves media cache by filename
	GetMediaCacheByFilename(ctx context.Context, filename string) (*WeblensFileImpl, error)

	// GetFileByContentID retrieves a file by its content ID
	GetFileByContentID(ctx context.Context, contentID string) (*WeblensFileImpl, error)

	// NewCacheFile creates a new cache file for media
	NewCacheFile(mediaID string, quality string, pageNum int) (*WeblensFileImpl, error)

	// DeleteCacheFile deletes a cache file
	DeleteCacheFile(file *WeblensFileImpl) error

	// GetFolderCover gets the cover for a folder
	// GetFolderCover(folder *WeblensFileImpl) (string, error)

	// SetFolderCover sets the cover for a folder
	// SetFolderCover(folderID string, coverID string) error

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
	NewZip(ctx context.Context, zipName string, owner *user_model.User) (*WeblensFileImpl, error)

	// GetZip retrieves a zip file by ID
	GetZip(ctx context.Context, id string) (*WeblensFileImpl, error)
}
