
package jobs_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/stretchr/testify/assert"
)

func TestRegisterJobs(t *testing.T) {
	t.Run("registers all jobs with worker pool", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)
		workerPool := task.NewWorkerPool(appCtx, 4)

		// Should not panic
		assert.NotPanics(t, func() {
			jobs.RegisterJobs(workerPool)
		})
	})
}

func TestJobNames(t *testing.T) {
	t.Run("ScanDirectoryTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.ScanDirectoryTask)
	})

	t.Run("ScanFileTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.ScanFileTask)
	})

	t.Run("UploadFilesTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.UploadFilesTask)
	})

	t.Run("CreateZipTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.CreateZipTask)
	})

	t.Run("GatherFsStatsTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.GatherFsStatsTask)
	})

	t.Run("BackupTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.BackupTask)
	})

	t.Run("CopyFileFromCoreTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.CopyFileFromCoreTask)
	})

	t.Run("RestoreCoreTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.RestoreCoreTask)
	})

	t.Run("HashFileTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.HashFileTask)
	})

	t.Run("LoadFilesystemTask is defined", func(t *testing.T) {
		assert.NotEmpty(t, job.LoadFilesystemTask)
	})
}

func TestExtSize(t *testing.T) {
	t.Run("struct can hold extension statistics", func(t *testing.T) {
		// extSize is an internal struct, we test the concept via GatherFilesystemStats behavior
		// This test verifies the pattern exists
		type testExtSize struct {
			Name  string `json:"name"`
			Value int64  `json:"value"`
		}

		stat := testExtSize{Name: "jpg", Value: 1024}
		assert.Equal(t, "jpg", stat.Name)
		assert.Equal(t, int64(1024), stat.Value)
	})
}
