// Package file provides interfaces and types for file operations in the Weblens system.
package file

import (
	"io"
	"os"

	"github.com/ethanrous/weblens/modules/wlfs"
)

// WeblensFile defines the interface for file operations in the Weblens system.
type WeblensFile interface {
	ID() string
	GetPortablePath() wlfs.Filepath
	GetContentID() string
	SetContentID(contentID string)
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
