package task

import (
	"context"

	"github.com/ethanrous/weblens/modules/task"
)

type Dispatcher interface {
	DispatchJob(ctx context.Context, jobName string, meta task.TaskMetadata) (*Task, error)
}
