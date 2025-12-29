// Package task provides task and worker pool management for the Weblens application.
package task

import (
	"context"

	"github.com/ethanrous/weblens/modules/task"
)

// Dispatcher provides an interface for dispatching jobs with associated metadata.
type Dispatcher interface {
	// DispatchJob creates and dispatches a new job with the specified name and metadata.
	DispatchJob(ctx context.Context, jobName string, meta task.Metadata) (*Task, error)
}
