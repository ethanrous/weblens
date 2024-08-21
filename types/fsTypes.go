package types

import (
	"fmt"
	"os"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/fileTree"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/websocket"
)

type FileTree interface {
	WeblensService[FileId, WeblensFile, FilesStore]

	Move(
		f, newParent WeblensFile, newFilename string, overwrite bool, event FileEvent,
		c ...websocket.BufferedBroadcasterAgent,
	) error

	Touch(parentFolder WeblensFile, newFileName string, detach bool, owner User, c ...websocket.BroadcasterAgent) (
		WeblensFile,
		error,
	)
	MkDir(parentFolder WeblensFile, newDirName string, event FileEvent, c ...websocket.BroadcasterAgent) (
		WeblensFile, error,
	)

	AttachFile(f WeblensFile, c ...websocket.BroadcasterAgent) error

	CreateHomeFolder(User) (WeblensFile, error)

	GetRoot() WeblensFile
	SetRoot(WeblensFile)
	GetJournal() fileTree.JournalService
	SetJournal(fileTree.JournalService)
	GenerateFileId(absPath string) FileId
	NewFile(parent WeblensFile, filename string, isDir bool, owner User) WeblensFile
	Size() int
	GetAllFiles() ([]WeblensFile, error)

	NewRoot(
		id FileId, filename, absPath string, owner User,
		parent WeblensFile,
	) (WeblensFile, error)
	SetDelDirectory(WeblensFile) error

	InitMediaRoot(websocket.BroadcasterAgent) error

	ResizeUp(WeblensFile, ...websocket.BroadcasterAgent) error
	ResizeDown(WeblensFile, ...websocket.BroadcasterAgent) error
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

	// FormatFileInfo(AccessMeta) (FileInfo, error)
	GetChildrenInfo(AccessMeta) []weblens.FileInfo

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
	Readable() (*os.File, error)
	Writeable() (*os.File, error)
	ReadAll() ([]byte, error)
	ReadDir() ([]WeblensFile, error)
	GetContentId() weblens.ContentId
	SetContentId(weblens.ContentId)

	GetShare() weblens.Share

	// SetShare UpdateShare(Share) error
	SetShare(weblens.Share) error
	RemoveShare(ShareId) error

	RecursiveMap(func(WeblensFile) error) error
	LeafMap(func(WeblensFile) error) error
	BubbleMap(func(WeblensFile) error) error
	IsParentOf(child WeblensFile) bool

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error

	LoadStat(casters ...websocket.BroadcasterAgent) (err error)
}

type WeblensFilepath interface {
	ToPortable() string
	String() string
	ToAbsPath() string
}



type FileStat struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Exists  bool      `json:"exists"`
}

var ErrNoFile = func(id FileId) error2.WErr {
	return error2.NewWeblensError(
		fmt.Sprintf(
			"cannot find file with id [%s]", id,
		),
	)
}

var ErrNoFileName = func(name string) error2.WErr {
	return error2.NewWeblensError(
		fmt.Sprintf(
			"cannot find file with name [%s]", name,
		),
	)
}
var ErrDirectoryRequired = error2.NewWeblensError("attempted to perform an action that requires a directory, but found regular file")
var ErrDirAlreadyExists = error2.NewWeblensError("directory already exists in destination location")
var ErrFileAlreadyExists = func(path string) error2.WErr {
	return error2.NewWeblensError("file already exists in destination location: " + path)
}
var ErrNoChildren = error2.NewWeblensError("file does not have any children")
var ErrChildAlreadyExists = error2.NewWeblensError("file already has the child being added")
var ErrDirNotAllowed = error2.NewWeblensError("attempted to perform action using a directory, where the action does not support directories")
var ErrIllegalFileMove = error2.NewWeblensError("tried to perform illegal file move")
var ErrWriteOnReadOnly = error2.NewWeblensError("tried to write to read-only file")
var ErrBadReadCount = error2.NewWeblensError("did not read expected number of bytes from file")
var ErrNoFileAccess = error2.NewWeblensError("user does not have access to file")
var ErrAlreadyWatching = error2.NewWeblensError("trying to watch directory that is already being watched")
var ErrBadTask = error2.NewWeblensError("did not get expected task id while trying to unlock file")
var ErrNoShare = error2.NewWeblensError("could not find share")
