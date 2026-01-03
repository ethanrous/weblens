package file_test

import (
	"context"
	"testing"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/task"
	task_model "github.com/ethanrous/weblens/models/task"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileService(t *testing.T) {
	t.Run("creates file service", func(t *testing.T) {
		ctx := context.Background()

		fs, err := file_service.NewFileService(ctx)

		require.NoError(t, err)
		assert.NotNil(t, fs)
	})
}

func TestFileService_Size(t *testing.T) {
	t.Run("returns -1 when tree not set", func(t *testing.T) {
		ctx := context.Background()
		fs, err := file_service.NewFileService(ctx)
		require.NoError(t, err)

		size := fs.Size("USERS")
		assert.Equal(t, int64(-1), size)
	})
}

func TestFileService_AddTask(t *testing.T) {
	ctx := context.Background()
	fs, err := file_service.NewFileService(ctx)
	require.NoError(t, err)

	t.Run("adds task to file", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		task := task_model.NewTestTask(
			"test-task-id",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, task)
		assert.NoError(t, err)

		// Verify task is associated with file
		tasks := fs.GetTasks(file)
		assert.Len(t, tasks, 1)
		assert.Equal(t, task, tasks[0])
	})

	t.Run("returns error when adding duplicate task", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file2.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		task := task_model.NewTestTask(
			"test-task-id-2",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, task)
		require.NoError(t, err)

		// Adding same task again should fail
		err = fs.AddTask(file, task)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileAlreadyHasTask)
	})
}

func TestFileService_RemoveTask(t *testing.T) {
	ctx := context.Background()
	fs, err := file_service.NewFileService(ctx)
	require.NoError(t, err)

	t.Run("removes task from file", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file3.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		task := task_model.NewTestTask(
			"test-task-id-3",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, task)
		require.NoError(t, err)

		err = fs.RemoveTask(file, task)
		assert.NoError(t, err)

		// Verify task is removed
		tasks := fs.GetTasks(file)
		assert.Empty(t, tasks)
	})

	t.Run("returns error when removing non-existent task", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file4.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		task := task_model.NewTestTask(
			"non-existent-task",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.RemoveTask(file, task)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNoTask)
	})
}

func TestFileService_GetTasks(t *testing.T) {
	ctx := context.Background()
	fs, err := file_service.NewFileService(ctx)
	require.NoError(t, err)

	t.Run("returns empty list for file without tasks", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file5.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		tasks := fs.GetTasks(file)
		assert.Empty(t, tasks)
	})

	t.Run("returns all tasks for file", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file6.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		task1 := task_model.NewTestTask(
			"task-1",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		task2 := task_model.NewTestTask(
			"task-2",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, task1)
		require.NoError(t, err)

		err = fs.AddTask(file, task2)
		require.NoError(t, err)

		tasks := fs.GetTasks(file)
		assert.Len(t, tasks, 2)
	})
}

func TestFolderCoverPair(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		pair := file_service.FolderCoverPair{
			FolderID:  "folder-123",
			ContentID: "content-456",
		}

		assert.Equal(t, "folder-123", pair.FolderID)
		assert.Equal(t, "content-456", pair.ContentID)
	})
}

func TestSkipJournalKey(t *testing.T) {
	t.Run("constant is defined", func(t *testing.T) {
		assert.Equal(t, "skipJournal", file_service.SkipJournalKey)
	})
}
