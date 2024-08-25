package werror

import (
	"fmt"

	"github.com/pkg/errors"
)

var ErrNoFile = errors.New("file not found")

var ErrNoFileId = func(id string) error {
	return Errorf("cannot find file with id [%s]", id)
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
var ErrDirAlreadyExists = errors.New("directory already exists in destination location")
var ErrFileAlreadyExists = New("file already exists in destination location")
var ErrNoChildren = errors.New("file does not have any children")
var ErrChildAlreadyExists = errors.New("file already has the child being added")
var ErrDirNotAllowed = errors.New("attempted to perform action using a directory, where the action does not support directories")
var ErrIllegalFileMove = errors.New("tried to perform illegal file move")
var ErrWriteOnReadOnly = errors.New("tried to write to read-only file")
var ErrBadReadCount = errors.New("did not read expected number of bytes from file")
var ErrAlreadyWatching = errors.New("trying to watch directory that is already being watched")
