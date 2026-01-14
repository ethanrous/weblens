package jobs_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/stretchr/testify/assert"
)

func TestRegisterJobs(t *testing.T) {
	t.Run("registers all jobs with worker pool", func(t *testing.T) {
		// logger := log.NewZeroLogger()
		// basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		// appCtx := ctxservice.NewAppContext(basicCtx)
		workerPool := task.NewWorkerPool(4)

		// Should not panic
		assert.NotPanics(t, func() {
			jobs.RegisterJobs(workerPool)
		})
	})
}
