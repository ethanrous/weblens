package werror

import (
	"errors"
	"fmt"
)

var ErrNonDisplayable = errors.New("attempt to process non-displayable file")
var ErrEmptyZip = errors.New("cannot create a zip with no files")
var ErrTaskExit = errors.New("task exit")
var ErrTaskError = errors.New("task generated an error")
var ErrTaskTimeout = errors.New("task timed out")
var ErrBadTaskType = errors.New("task metadata type is not supported on attempted operation")
var ErrBadCaster = errors.New("task was given the wrong type of caster")
var ErrChildTaskFailed = errors.New("a task spawned by this task has failed")

func ErrBadJobMetadata(jobName string, field string) error {
	return &withStack{
		err:   errors.New(fmt.Sprintf("Bad %s metadata, missing %s", jobName, field)),
		stack: callers(),
	}
}