package file

import (
	"io"
	"os"

	"github.com/ethanrous/weblens/modules/fs"
)

type WeblensFile interface {
	ID() string
	GetPortablePath() fs.Filepath
	GetContentId() string
	SetContentId(contentId string)
	Size() int64
	IsDir() bool
	Exists() bool
	// Freeze() WeblensFile
	// GetParent() WeblensFile
	Write(data []byte) (int, error)
	Readable() (io.Reader, error)
	// GetChildren() []WeblensFile
	CreateSelf() error
	Remove() error
	LoadStat() (os.FileInfo, error)
	AbsPath() string
}
