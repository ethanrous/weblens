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
	IsDisplayable() bool

	// GetMedia() (Media, error)
	// SetMedia(Media) error
	// ClearMedia()

	Copy() WeblensFile
	GetParent() WeblensFile
	GetChildren() []WeblensFile
	GetChildrenInfo(AccessMeta) []FileInfo
	GetChild(childName string) (WeblensFile, error)
	AddChild(child WeblensFile) error
	IsReadOnly() bool

	AddTask(Task)
	GetTask() Task
	RemoveTask(TaskId) error

	CreateSelf() error
	Write([]byte) error
	WriteAt([]byte, int64) error
	Read() (*os.File, error)
	ReadAll() ([]byte, error)
	ReadDir() error
	GetContentId() ContentId

	GetShare(ShareId) (Share, error)
	GetShares() []Share
	UpdateShare(Share) error
	AppendShare(Share)
	RemoveShare(ShareId) error

	RecursiveMap(FileMapFunc) error
	LeafMap(FileMapFunc) error
	BubbleMap(FileMapFunc) error

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
	MarshalArchive() map[string]any
}

type FileId string

func (fId FileId) String() string {
	return string(fId)
}

type FileMapFunc func(WeblensFile) error

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
	PastFile         bool      `json:"pastFile"`
}
