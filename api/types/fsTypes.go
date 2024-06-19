package types

import (
	"os"
	"time"
)

type FileTree interface {
	BaseService[FileId, WeblensFile]

	Move(f, newParent WeblensFile, newFilename string, overwrite bool, c ...BufferedBroadcasterAgent) error

	Touch(parentFolder WeblensFile, newFileName string, detach bool, owner User, c ...BroadcasterAgent) (
		WeblensFile,
		error,
	)
	MkDir(parentFolder WeblensFile, newDirName string, c ...BroadcasterAgent) (WeblensFile, error)

	AttachFile(f WeblensFile, c ...BroadcasterAgent) error

	GetRoot() WeblensFile
	SetRoot(WeblensFile)
	GetJournal() JournalService
	SetJournal(JournalService)
	GenerateFileId(absPath string) FileId
	NewFile(parent WeblensFile, filename string, isDir bool) WeblensFile
	Size() int

	AddRoot(r WeblensFile) error
	NewRoot(
		id FileId, filename, absPath string, owner User,
		parent WeblensFile,
	) (WeblensFile, error)
	SetDelDirectory(WeblensFile) error

	ResizeUp(WeblensFile, ...BroadcasterAgent) error
	ResizeDown(WeblensFile, ...BroadcasterAgent) error
}

type WeblensFile interface {
	ID() FileId
	GetTree() FileTree
	IsDir() bool
	Size() (int64, error)
	Filename() string
	GetAbsPath() string
	Owner() User
	SetOwner(User)
	Exists() bool
	ModTime() time.Time

	FormatFileInfo(AccessMeta) (FileInfo, error)
	GetChildrenInfo(AccessMeta) []FileInfo

	IsDisplayable() bool

	Copy() WeblensFile
	GetParent() WeblensFile
	GetChildren() []WeblensFile
	GetChild(childName string) (WeblensFile, error)
	AddChild(child WeblensFile) error
	IsReadOnly() bool

	AddTask(Task)
	GetTask() Task
	RemoveTask(TaskId) error

	SetWatching() error

	CreateSelf() error
	Write([]byte) error
	WriteAt([]byte, int64) error
	Read() (*os.File, error)
	ReadAll() ([]byte, error)
	ReadDir() ([]WeblensFile, error)
	GetContentId() ContentId
	SetContentId(ContentId)

	GetShare(ShareId) (Share, error)
	GetShares() []Share
	// UpdateShare(Share) error
	AppendShare(Share)
	RemoveShare(ShareId) error

	RecursiveMap(FileMapFunc) error
	LeafMap(FileMapFunc) error
	BubbleMap(FileMapFunc) error

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
	MarshalArchive() map[string]any
	LoadStat(casters ...BroadcasterAgent) (err error)
}

type FileId string

func (fId FileId) String() string {
	return string(fId)
}

type FileMapFunc func(WeblensFile) error

// FileInfo is a structure for safely sending file information to the client
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

var ErrNoFile = NewWeblensError("file does not exist")
var ErrDirectoryRequired = NewWeblensError("attempted to perform an action that requires a directory, but found regular file")
var ErrDirAlreadyExists = NewWeblensError("directory already exists in destination location")
var ErrFileAlreadyExists = NewWeblensError("file already exists in destination location")
var ErrChildAlreadyExists = NewWeblensError("file already has the child being added")
var ErrDirNotAllowed = NewWeblensError("attempted to perform action using a directory, where the action does not support directories")
var ErrIllegalFileMove = NewWeblensError("tried to perform illegal file move")
var ErrWriteOnReadOnly = NewWeblensError("tried to write to read-only file")
var ErrBadReadCount = NewWeblensError("did not read expected number of bytes from file")
var ErrNoFileAccess = NewWeblensError("user does not have access to file")
var ErrAlreadyWatching = NewWeblensError("trying to watch directory that is already being watched")
var ErrBadTask = NewWeblensError("did not get expected task id while trying to unlock file")
var ErrNoShare = NewWeblensError("could not find share")
