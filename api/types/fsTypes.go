package types

import (
	"os"
	"time"
)

type WeblensFile interface {
	Id() FileId
	IsDir() bool
	Size() (int64, error)
	Filename() string
	GetAbsPath() string
	Owner() User
	Exists() bool
	ModTime() time.Time

	FormatFileInfo(AccessMeta) (FileInfo, error)
	IsDisplayable() (bool, error)

	GetMedia() (Media, error)
	SetMedia(Media) error
	ClearMedia()

	Copy() WeblensFile
	GetParent() WeblensFile
	GetChildren() []WeblensFile
	GetChildrenInfo(AccessMeta) []FileInfo
	AddChild(WeblensFile)
	IsReadOnly() bool

	AddTask(Task)
	GetTasks() []Task
	RemoveTask(TaskId) bool

	CreateSelf() error
	Write([]byte) error
	WriteAt([]byte, int64) error
	Read() (*os.File, error)
	ReadAll() ([]byte, error)
	ReadDir() error

	GetShare(ShareId) (Share, error)
	GetShares() []Share
	UpdateShare(Share) error
	AppendShare(Share)
	RemoveShare(ShareId) error

	RecursiveMap(func(WeblensFile))
	LeafMap(func(WeblensFile))
	BubbleMap(func(WeblensFile))
}

type FileId string

func (fId FileId) String() string {
	return string(fId)
}

// Structure for safely sending file information to the client
type FileInfo struct {
	Id FileId `json:"id"`

	// If the media has been loaded into the database, only if it should be.
	// If media is not required to be imported, this will be set true
	Imported bool `json:"imported"`

	// If the content of the file can be displayed visually.
	// Say the file is a jpg, mov, arw, etc. and not a zip,
	// txt, doc, directory etc.
	Displayable bool `json:"displayable"`

	IsDir            bool      `json:"isDir"`
	Modifiable       bool      `json:"modifiable"`
	Size             int64     `json:"size"`
	ModTime          time.Time `json:"modTime"`
	Filename         string    `json:"filename"`
	ParentFolderId   FileId    `json:"parentFolderId"`
	MediaData        Media     `json:"mediaData"`
	FileFriendlyName string    `json:"fileFriendlyName"`
	Owner            Username  `json:"owner"`
	PathFromHome     string    `json:"pathFromHome"`
	Shares           []Share   `json:"shares"`
	Children         []FileId  `json:"children"`
}
