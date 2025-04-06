package file

import (
	"fmt"
	"os"

	"github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/fs"
)

const (
	UsersTreeKey   = "USERS"
	RestoreTreeKey = "RESTORE"
	CachesTreeKey  = "CACHES"

	UserTrashDirName = ".user_trash"
	ThumbsDirName    = "thumbs"
)

var UsersRootPath = fs.Filepath{RootAlias: UsersTreeKey}
var ThumbsDirPath = fs.Filepath{RootAlias: CachesTreeKey, RelPath: ThumbsDirName}
var RestoreDirPath = fs.Filepath{RootAlias: RestoreTreeKey}

func isCacheFile(filepath fs.Filepath) bool {
	return filepath.RootAlias == CachesTreeKey
}

func cacheFilename(mId, quality string, pageNum int) string {
	var pageNumStr string
	if pageNum > 1 && quality == string(media_model.HighRes) {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}
	filename := fmt.Sprintf("%s-%s%s.cache", mId, quality, pageNumStr)

	return filename
}

func touch(filepath fs.Filepath) (f *file.WeblensFileImpl, err error) {
	osf, err := os.Create(filepath.ToAbsolute())
	if err != nil {
		return
	}

	err = osf.Close()
	if err != nil {
		return
	}

	f = file.NewWeblensFile(filepath)
	return
}

func mkdir(filepath fs.Filepath) (f *file.WeblensFileImpl, err error) {
	err = os.Mkdir(filepath.ToAbsolute(), os.ModePerm)
	if err != nil {
		return
	}

	f = file.NewWeblensFile(filepath)
	return
}

func exists(filepath fs.Filepath) bool {
	_, err := os.Stat(filepath.ToAbsolute())
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		panic(err)
	}

	return true
}

func rename(oldPath, newPath fs.Filepath) error {
	return os.Rename(oldPath.ToAbsolute(), newPath.ToAbsolute())
}

func remove(filepath fs.Filepath) error {
	return os.Remove(filepath.ToAbsolute())
}
