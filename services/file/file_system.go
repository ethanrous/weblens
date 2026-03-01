package file

import (
	"os"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlfs"
)

func isCacheFile(filepath wlfs.Filepath) bool {
	return filepath.RootAlias == file_model.CachesTreeKey
}

func touch(filepath wlfs.Filepath) (f *file_model.WeblensFileImpl, err error) {
	osf, err := os.Create(filepath.ToAbsolute())
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	err = osf.Close()
	if err != nil {
		return nil, wlerrors.WithStack(err)
	}

	f = file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})

	return
}

func mkdir(filepath wlfs.Filepath) (f *file_model.WeblensFileImpl, err error) {
	err = os.Mkdir(filepath.ToAbsolute(), os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return nil, wlerrors.Wrap(file_model.ErrDirectoryAlreadyExists, filepath.ToAbsolute())
		}

		return
	}

	f = file_model.NewWeblensFile(file_model.NewFileOptions{Path: filepath})

	return
}

func exists(filepath wlfs.Filepath) bool {
	_, err := os.Stat(filepath.ToAbsolute())
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		panic(err)
	}

	return true
}

func rename(oldPath, newPath wlfs.Filepath) error {
	return wlerrors.WithStack(os.Rename(oldPath.ToAbsolute(), newPath.ToAbsolute()))
}

func remove(filepath wlfs.Filepath) error {
	return wlerrors.WithStack(os.RemoveAll(filepath.ToAbsolute()))
}

func getChildFilepaths(filepath wlfs.Filepath) ([]wlfs.Filepath, error) {
	childs, err := os.ReadDir(filepath.ToAbsolute())
	if err != nil {
		return nil, err
	}

	childPaths := make([]wlfs.Filepath, 0, len(childs))

	for _, child := range childs {
		childPath := filepath.Child(child.Name(), child.IsDir())
		childPaths = append(childPaths, childPath)
	}

	return childPaths, nil
}
