package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/task"
)

type DispatcherContext interface {
	context.Context
	DispatchJob(string, task.TaskMetadata, task.Pool) (task.Task, error)
}
