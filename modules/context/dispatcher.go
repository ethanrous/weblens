package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/task"
)

// DispatcherContext provides the ability to dispatch jobs as tasks.
type DispatcherContext interface {
	context.Context
	DispatchJob(jobName string, metadata task.Metadata, pool task.Pool) (task.Task, error)
}
