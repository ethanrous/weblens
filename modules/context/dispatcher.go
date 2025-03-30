package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/task"
)

type DispatcherContext interface {
	context.Context
	DispatchJob(string, any) (task.Task, error)
}
