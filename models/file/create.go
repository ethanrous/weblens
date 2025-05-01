package file

import (
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/option"
)

type NewFileOptions struct {
	Path file_system.Filepath

	FileId    string
	ContentId string

	IsPastFile bool

	Size int64

	// Parent *WeblensFileImpl
}

func NewWeblensFile(params NewFileOptions) *WeblensFileImpl {
	if params.Path.IsZero() {
		panic("Path cannot be empty")
	}

	f := &WeblensFileImpl{
		portablePath: params.Path,
		childrenMap:  make(map[string]*WeblensFileImpl),
		isDir:        option.Of(params.Path.IsDir()),

		id:        params.FileId,
		contentId: params.ContentId,
		// parent:    params.Parent,
		pastFile: params.IsPastFile,
	}

	if params.Size > 0 {
		f.size.Store(params.Size)
	} else {
		f.size.Store(-1)
	}

	return f
}
