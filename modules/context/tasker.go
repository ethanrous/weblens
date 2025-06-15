package context

import "github.com/ethanrous/weblens/modules/task"

type Tasker interface {
	DispatchJob(ctx LoggerContext, jobName string, meta task.TaskMetadata, pool task.Pool) (task.Task, error)
}
