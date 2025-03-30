package task

import "context"

type Dispatcher interface {
	DispatchJob(ctx context.Context, jobName string, meta TaskMetadata) (*Task, error)
}
