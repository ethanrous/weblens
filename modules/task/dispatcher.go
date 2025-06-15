package task

import "context"

type Dispatcher interface {
	DispatchJob(context.Context, string, TaskMetadata, Pool) (Task, error)
}
