package file

import (
	"context"
	"strings"

	"github.com/ethanrous/weblens/modules/fs"
	"github.com/pkg/errors"
)

const (
	UsersTreeKey   = "USERS"
	RestoreTreeKey = "RESTORE"
	CachesTreeKey  = "CACHES"

	UserTrashDirName = ".user_trash"
	ThumbsDirName    = "thumbs"
)

var UsersRootPath = fs.Filepath{RootAlias: UsersTreeKey}
var CacheRootPath = fs.Filepath{RootAlias: CachesTreeKey}
var ThumbsDirPath = fs.Filepath{RootAlias: CachesTreeKey, RelPath: ThumbsDirName}
var RestoreDirPath = fs.Filepath{RootAlias: RestoreTreeKey}

func GetFileOwnerName(ctx context.Context, file *WeblensFileImpl) (string, error) {
	if file == nil {
		return "", errors.WithStack(ErrNilFile)
	}

	return GetFileOwnerNameFromPath(ctx, file.GetPortablePath())
}

func GetFileOwnerNameFromPath(ctx context.Context, portable fs.Filepath) (string, error) {
	if portable.RootName() != UsersTreeKey {
		return "", errors.New("trying to get owner of file not in USERS tree")
	}

	slashIndex := strings.Index(portable.RelativePath(), "/")

	var username string
	if slashIndex == -1 {
		username = portable.RelativePath()
	} else {
		username = portable.RelativePath()[:slashIndex]
	}

	if username == "" {
		return "", errors.Errorf("could not find username in file path [%s]", portable.RelativePath())
	}

	return username, nil
}

func IsFileInTrash(f *WeblensFileImpl) bool {
	return strings.Contains(f.GetPortablePath().RelativePath(), UserTrashDirName)
}
