package file

import (
	"context"
	"strings"

	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/wlerrors"
)

const (
	// UsersTreeKey identifies the root of the users file tree.
	UsersTreeKey = "USERS"
	// RestoreTreeKey identifies the root of the restore file tree.
	RestoreTreeKey = "RESTORE"
	// CachesTreeKey identifies the root of the caches file tree.
	CachesTreeKey = "CACHES"
	// BackupTreeKey identifies the root of the backup file tree.
	BackupTreeKey = "BACKUP"

	// UserTrashDirName is the directory name for user trash folders.
	UserTrashDirName = ".user_trash"
	// ThumbsDirName is the directory name for thumbnail storage.
	ThumbsDirName = "thumbs/"
	// ZipsDirName is the directory name for zip file storage.
	ZipsDirName = "zips/"
)

// UsersRootPath is the filepath to the users file tree root.
var UsersRootPath = fs.Filepath{RootAlias: UsersTreeKey}

// BackupRootPath is the filepath to the backup file tree root.
var BackupRootPath = fs.Filepath{RootAlias: BackupTreeKey}

// CacheRootPath is the filepath to the cache file tree root.
var CacheRootPath = fs.Filepath{RootAlias: CachesTreeKey}

// ZipsDirPath is the filepath to the zips storage directory.
var ZipsDirPath = fs.Filepath{RootAlias: CachesTreeKey, RelPath: ZipsDirName}

// ThumbsDirPath is the filepath to the thumbnails storage directory.
var ThumbsDirPath = fs.Filepath{RootAlias: CachesTreeKey, RelPath: ThumbsDirName}

// RestoreDirPath is the filepath to the restore directory.
var RestoreDirPath = fs.Filepath{RootAlias: RestoreTreeKey}

// GetFileOwnerName retrieves the username of the file owner from a file instance.
func GetFileOwnerName(ctx context.Context, file *WeblensFileImpl) (string, error) {
	if file == nil {
		return "", wlerrors.WithStack(ErrNilFile)
	}

	return GetFileOwnerNameFromPath(ctx, file.GetPortablePath())
}

// GetFileOwnerNameFromPath extracts the username of the file owner from a portable filepath.
func GetFileOwnerNameFromPath(_ context.Context, portable fs.Filepath) (string, error) {
	if portable.RootName() == BackupTreeKey {
		return "", nil
	}
	// slashIndex := strings.Index(portable.RelativePath(), "/")
	// if slashIndex == -1 {
	// 	portable = fs.BuildFilePath(UsersTreeKey, portable.RelativePath()[slashIndex:])
	// }

	if portable.RootName() != UsersTreeKey {
		return "", wlerrors.Errorf("trying to get owner of file not in USERS tree: [%s]", portable)
	}

	slashIndex := strings.Index(portable.RelativePath(), "/")

	var username string
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}

	if username == "" {
		return "", wlerrors.Errorf("could not find username in file path [%s]", portable.RelativePath())
	}

	return username, nil
}

// IsFileInTrash checks whether a file is located in the user trash directory.
func IsFileInTrash(f *WeblensFileImpl) bool {
	return strings.Contains(f.GetPortablePath().RelativePath(), UserTrashDirName)
}
