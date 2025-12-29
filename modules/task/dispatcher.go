// Package task provides task dispatching and execution management.
package task

import "context"

// Dispatcher creates and dispatches tasks to a pool for execution.
type Dispatcher interface {
	DispatchJob(ctx context.Context, jobName string, metadata Metadata, pool Pool) (Task, error)
}
