package file

import (
	"fmt"
	"os"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/pkg/errors"
)

func isCacheFile(filepath fs.Filepath) bool {
	return filepath.RootAlias == file_model.CachesTreeKey
}

func getCacheFilename(mId, quality string, pageNum int) string {
	var pageNumStr string
	if pageNum > 1 && quality == string(media_model.HighRes) {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}

	return fmt.Sprintf("%s-%s%s.cache", mId, quality, pageNumStr)
}

func touch(filepath fs.Filepath) (f *file_model.WeblensFileImpl, err error) {
	osf, err := os.Create(filepath.ToAbsolute())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = osf.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	f = file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})

	return
}

func mkdir(filepath fs.Filepath) (f *file_model.WeblensFileImpl, err error) {
	err = os.Mkdir(filepath.ToAbsolute(), os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return nil, errors.Wrap(file_model.ErrDirectoryAlreadyExists, filepath.ToAbsolute())
		}

		return
	}

	f = file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})

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
	return errors.WithStack(os.Rename(oldPath.ToAbsolute(), newPath.ToAbsolute()))
}

func remove(filepath fs.Filepath) error {
	return errors.WithStack(os.RemoveAll(filepath.ToAbsolute()))
}

func getChildFilepaths(filepath fs.Filepath) ([]fs.Filepath, error) {
	childs, err := os.ReadDir(filepath.ToAbsolute())
	if err != nil {
		return nil, err
	}

	childPaths := make([]fs.Filepath, 0, len(childs))

	for _, child := range childs {
		childPath := filepath.Child(child.Name(), child.IsDir())
		childPaths = append(childPaths, childPath)
	}

	return childPaths, nil
}
