package types

import (
	"fmt"
	"os"
	"time"
)

type FileTree interface {
	WeblensService[FileId, WeblensFile, FilesStore]

	Move(
		f, newParent WeblensFile, newFilename string, overwrite bool, event FileEvent,
		c ...BufferedBroadcasterAgent,
	) error

	Touch(parentFolder WeblensFile, newFileName string, detach bool, owner User, c ...BroadcasterAgent) (
		WeblensFile,
		error,
	)
	MkDir(parentFolder WeblensFile, newDirName string, event FileEvent, c ...BroadcasterAgent) (WeblensFile, error)

	AttachFile(f WeblensFile, c ...BroadcasterAgent) error

	GetRoot() WeblensFile
	SetRoot(WeblensFile)
	GetJournal() JournalService
	SetJournal(JournalService)
	GenerateFileId(absPath string) FileId
	NewFile(parent WeblensFile, filename string, isDir bool, owner User) WeblensFile
	Size() int
	GetAllFiles() ([]WeblensFile, error)

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
	GetPortablePath() WeblensFilepath

	Owner() User
	SetOwner(User)
	Exists() bool
	ModTime() time.Time

	FormatFileInfo(AccessMeta) (FileInfo, error)
	GetChildrenInfo(AccessMeta) []FileInfo

	IsDisplayable() bool
	IsDetached() bool
	IsReadOnly() bool

	Copy() WeblensFile
	GetParent() WeblensFile
	GetChildren() []WeblensFile
	GetChild(childName string) (WeblensFile, error)
	AddChild(child WeblensFile) error

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

	GetShare() Share

	// SetShare UpdateShare(Share) error
	SetShare(Share) error
	RemoveShare(ShareId) error

	RecursiveMap(FileMapFunc) error
	LeafMap(FileMapFunc) error
	BubbleMap(FileMapFunc) error
	IsParentOf(child WeblensFile) bool

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
	// MarshalArchive() map[string]any
	LoadStat(casters ...BroadcasterAgent) (err error)
}

type WeblensFilepath interface {
	ToPortable() string
	String() string
	ToAbsPath() string
}

type TrashEntry struct {
	OrigParent   FileId `bson:"originalParentId"`
	OrigFilename string `bson:"originalFilename"`
	TrashFileId  FileId `bson:"trashFileId"`
}

type FileStat struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Exists  bool      `json:"exists"`
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

	IsDir          bool     `json:"isDir"`
	Modifiable     bool     `json:"modifiable"`
	Size           int64    `json:"size"`
	ModTime        int64    `json:"modTime"`
	Filename       string   `json:"filename"`
	ParentFolderId FileId   `json:"parentFolderId"`
	MediaData      Media    `json:"mediaData"`
	Owner          Username `json:"owner"`
	PathFromHome   string   `json:"pathFromHome"`
	ShareId        ShareId  `json:"shareId"`
	Children       []FileId `json:"children"`
	PastFile       bool     `json:"pastFile"`
}

var ErrNoFile = func(id FileId) WeblensError {
	return NewWeblensError(
		fmt.Sprintf(
			"cannot find file with id [%s]", id,
		),
	)
}

var ErrNoFileName = func(name string) WeblensError {
	return NewWeblensError(
		fmt.Sprintf(
			"cannot find file with name [%s]", name,
		),
	)
}
var ErrDirectoryRequired = NewWeblensError("attempted to perform an action that requires a directory, but found regular file")
var ErrDirAlreadyExists = NewWeblensError("directory already exists in destination location")
var ErrFileAlreadyExists = func(path string) WeblensError { return NewWeblensError("file already exists in destination location: " + path) }
var ErrNoChildren = NewWeblensError("file does not have any children")
var ErrChildAlreadyExists = NewWeblensError("file already has the child being added")
var ErrDirNotAllowed = NewWeblensError("attempted to perform action using a directory, where the action does not support directories")
var ErrIllegalFileMove = NewWeblensError("tried to perform illegal file move")
var ErrWriteOnReadOnly = NewWeblensError("tried to write to read-only file")
var ErrBadReadCount = NewWeblensError("did not read expected number of bytes from file")
var ErrNoFileAccess = NewWeblensError("user does not have access to file")
var ErrAlreadyWatching = NewWeblensError("trying to watch directory that is already being watched")
var ErrBadTask = NewWeblensError("did not get expected task id while trying to unlock file")
var ErrNoShare = NewWeblensError("could not find share")
