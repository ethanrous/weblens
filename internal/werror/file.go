package werror

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrNoFileName struct {
	Err  error
	Name string
}

func (e *ErrNoFileName) Error() string {
	return fmt.Sprintf("cannot find file with name [%s]", e.Name)
}

func (e *ErrNoFileName) Unwrap() error {
	return e.Err
}

var NewErrNoFileName = func(name string) error {
	return &ErrNoFileName{
		Name: name,
		Err:  ErrNoFile,
	}
}

var fileNotFound = errors.New("file not found or user does not have access to it")

var ErrNoFile = clientSafeErr{
	realError:  errors.New("file does not exist"),
	safeErr:    fileNotFound,
	statusCode: http.StatusNotFound,
}

var ErrNoFileAccess = clientSafeErr{
	realError:  errors.New("user does not have access to file"),
	safeErr:    fileNotFound,
	statusCode: http.StatusNotFound,
}

var ErrDirectoryRequired = errors.New(
	"attempted to perform an action that requires a directory, " +
		"but found regular file",
)

var ErrDirAlreadyExists = clientSafeErr{
	safeErr:    errors.New("directory already exists in destination location"),
	statusCode: http.StatusConflict,
}

var ErrFileAlreadyExists = clientSafeErr{
	safeErr:    errors.New("file already exists in destination location"),
	statusCode: http.StatusConflict,
}

var ErrNilFile = errors.New("file is required but is nil")
var ErrFilenameRequired = errors.New("filename is required but is empty")
var ErrEmptyMove = errors.New("refusing to perform move with same filename and same parent")
var ErrNoChildren = errors.New("file does not have any children")
var ErrDirNotAllowed = errors.New("attempted to perform action using a directory, where the action does not support directories")
var ErrBadReadCount = errors.New("did not read expected number of bytes from file")
var ErrAlreadyWatching = errors.New("trying to watch directory that is already being watched")
var ErrFileAlreadyHasTask = errors.New("file already has a task")
var ErrFileNoTask = errors.New("file does not have task")
var ErrNoContentId = errors.New("file does not have a content id")

var ErrNoFileTree = clientSafeErr{
	realError:  errors.New("filetree does not exist"),
	safeErr:    errors.New("trying to get a filetree that does not exist"),
	statusCode: http.StatusNotFound,
}

var ErrJournalServerMismatch = errors.New("journal serverId does not match the lifetime serverId")
