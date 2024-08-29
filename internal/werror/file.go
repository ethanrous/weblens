package werror

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrNoFile = &clientSafeErr{
	safeErr: errors.New("file does not exist or user does not have access to it"),
	statusCode: http.StatusNotFound,
}

var ErrNoFileId = func(id string) error {
	err := *ErrNoFile
	err.realError = Errorf("cannot find file with id [%s]", id)
	return &err
}

var ErrNoFileName = func(name string) error {
	return errors.New(
		fmt.Sprintf(
			"cannot find file with name [%s]", name,
		),
	)
}

var ErrDirectoryRequired = errors.New(
	"attempted to perform an action that requires a directory, " +
		"but found regular file",
)

var ErrDirAlreadyExists = &clientSafeErr{
	safeErr:    errors.New("directory already exists in destination location"),
	statusCode: http.StatusConflict,
}

var ErrFileAlreadyExists = &clientSafeErr{
	safeErr:    errors.New("file already exists in destination location"),
	statusCode: http.StatusConflict,
}

var ErrFileRequired = errors.New("file is required but is nil")
var ErrFilenameRequired = errors.New("filename is required but is empty")

var ErrEmptyMove = errors.New("refusing to perform move with same filename and same parent")

var ErrNoChildren = errors.New("file does not have any children")
var ErrDirNotAllowed = errors.New("attempted to perform action using a directory, where the action does not support directories")
var ErrBadReadCount = errors.New("did not read expected number of bytes from file")
var ErrAlreadyWatching = errors.New("trying to watch directory that is already being watched")
var ErrFileAlreadyHasTask = errors.New("file already has a task")
var ErrFileNoTask = errors.New("file does not have task")
