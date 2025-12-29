package context

import "github.com/ethanrous/weblens/modules/task"

// Tasker dispatches jobs as tasks with logging context.
type Tasker interface {
	DispatchJob(ctx LoggerContext, jobName string, meta task.Metadata, pool task.Pool) (task.Task, error)
}
