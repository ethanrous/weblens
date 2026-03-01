package file //nolint:testpackage

import (
	"context"
	"testing"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/task"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/wlog"
	ctxservice "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext() context.Context {
	logger := wlog.NewZeroLogger()
	basicCtx := ctxservice.NewBasicContext(context.Background(), logger)

	return ctxservice.NewAppContext(basicCtx)
}

func TestNewFileService(t *testing.T) {
	t.Run("creates file service", func(t *testing.T) {
		ctx := newTestContext()

		fs, err := NewFileService(ctx)

		require.NoError(t, err)
		assert.NotNil(t, fs)
	})
}

func TestFileService_Size(t *testing.T) {
	t.Run("returns -1 when tree not set", func(t *testing.T) {
		ctx := newTestContext()
		fs, err := NewFileService(ctx)
		require.NoError(t, err)

		size := fs.Size("USERS")
		assert.Equal(t, int64(-1), size)
	})
}

func TestFileService_AddTask(t *testing.T) {
	ctx := newTestContext()
	fs, err := NewFileService(ctx)
	require.NoError(t, err)

	t.Run("adds task to file", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		testTask := task_model.NewTestTask(
			"test-task-id",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, testTask)
		assert.NoError(t, err)

		// Verify task is associated with file
		tasks := fs.GetTasks(file)
		assert.Len(t, tasks, 1)
		assert.Equal(t, testTask, tasks[0])
	})

	t.Run("returns error when adding duplicate task", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file2.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		testTask := task_model.NewTestTask(
			"test-task-id-2",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, testTask)
		require.NoError(t, err)

		// Adding same task again should fail
		err = fs.AddTask(file, testTask)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileAlreadyHasTask)
	})
}

func TestFileService_RemoveTask(t *testing.T) {
	ctx := newTestContext()
	fs, err := NewFileService(ctx)
	require.NoError(t, err)

	t.Run("removes task from file", func(t *testing.T) {
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("testuser/file3.txt", false),
			MemOnly:    true,
			GenerateID: true,
		})

		testTask := task_model.NewTestTask(
			"test-task-id-3",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.AddTask(file, testTask)
		require.NoError(t, err)

		err = fs.RemoveTask(file, testTask)
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

		testTask := task_model.NewTestTask(
			"non-existent-task",
			"TestJob",
			1,
			task.TaskSuccess,
			task.Result{},
			time.Now(),
		)

		err := fs.RemoveTask(file, testTask)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNoTask)
	})
}

func TestFileService_GetTasks(t *testing.T) {
	ctx := newTestContext()
	fs, err := NewFileService(ctx)
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

func TestFileService_MoveFiles_Integration(t *testing.T) {
	t.Run("moves single file to new folder", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Get user home directory
		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create source file and destination folder
		sourceFile := createTestFile(t, ctx, fs, userHome, "source.txt", []byte("test content"))
		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")

		// Move the file
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{sourceFile}, destFolder)
		require.NoError(t, err)

		// Verify: file at new location
		expectedPath := file_model.UsersRootPath.Child("testuser/dest/source.txt", false)
		movedFile, err := fs.GetFileByFilepath(ctx, expectedPath)
		require.NoError(t, err)
		assert.Equal(t, sourceFile.ID(), movedFile.ID())
		assertFileExistsOnDisk(t, expectedPath)

		// Verify: old location doesn't exist
		oldPath := file_model.UsersRootPath.Child("testuser/source.txt", false)
		assertFileNotExistsOnDisk(t, oldPath)

		// Verify: parent-child relationships updated
		assert.Equal(t, destFolder.ID(), movedFile.GetParent().ID())
	})

	t.Run("moves multiple files to destination", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Get user home directory
		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create 3 source files and destination folder
		file1 := createTestFile(t, ctx, fs, userHome, "file1.txt", []byte("content 1"))
		file2 := createTestFile(t, ctx, fs, userHome, "file2.txt", []byte("content 2"))
		file3 := createTestFile(t, ctx, fs, userHome, "file3.txt", []byte("content 3"))
		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")

		// Move all files
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{file1, file2, file3}, destFolder)
		require.NoError(t, err)

		// Verify: all files moved
		for _, filename := range []string{"file1.txt", "file2.txt", "file3.txt"} {
			expectedPath := file_model.UsersRootPath.Child("testuser/dest/"+filename, false)
			_, err := fs.GetFileByFilepath(ctx, expectedPath)
			assert.NoError(t, err, "file %s should exist at new location", filename)
			assertFileExistsOnDisk(t, expectedPath)

			// Verify old location doesn't exist
			oldPath := file_model.UsersRootPath.Child("testuser/"+filename, false)
			assertFileNotExistsOnDisk(t, oldPath)
		}
	})

	t.Run("moves folder with children recursively", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Get user home directory
		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create folder/subfolder/file.txt structure
		sourceFolder := createTestFolder(t, ctx, fs, userHome, "folder")
		subfolder := createTestFolder(t, ctx, fs, sourceFolder, "subfolder")
		_ = createTestFile(t, ctx, fs, subfolder, "file.txt", []byte("nested content"))

		// Create destination
		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")

		// Move the folder
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{sourceFolder}, destFolder)
		require.NoError(t, err)

		// Verify: all descendant paths updated
		expectedFolderPath := file_model.UsersRootPath.Child("testuser/dest/folder", true)
		expectedSubfolderPath := file_model.UsersRootPath.Child("testuser/dest/folder/subfolder", true)
		expectedFilePath := file_model.UsersRootPath.Child("testuser/dest/folder/subfolder/file.txt", false)

		movedFolder, err := fs.GetFileByFilepath(ctx, expectedFolderPath)
		require.NoError(t, err)
		assert.Equal(t, sourceFolder.ID(), movedFolder.ID())
		assertFileExistsOnDisk(t, expectedFolderPath)

		movedSubfolder, err := fs.GetFileByFilepath(ctx, expectedSubfolderPath)
		require.NoError(t, err)
		assertFileExistsOnDisk(t, expectedSubfolderPath)

		movedFile, err := fs.GetFileByFilepath(ctx, expectedFilePath)
		require.NoError(t, err)
		assertFileExistsOnDisk(t, expectedFilePath)

		// Verify filesystem matches - the file still has correct parent chain
		assert.Equal(t, movedSubfolder.ID(), movedFile.GetParent().ID())
		assert.Equal(t, movedFolder.ID(), movedSubfolder.GetParent().ID())
	})
}

func TestFileService_RenameFile_Integration(t *testing.T) {
	t.Run("renames file successfully", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "old.txt", []byte("content"))
		oldPath := file.GetPortablePath()

		err = fs.RenameFile(ctx, file, "new.txt")
		require.NoError(t, err)

		// Verify: file at new path
		newPath := file_model.UsersRootPath.Child("testuser/new.txt", false)
		renamedFile, err := fs.GetFileByFilepath(ctx, newPath)
		require.NoError(t, err)
		assert.Equal(t, file.ID(), renamedFile.ID())
		assertFileExistsOnDisk(t, newPath)

		// Verify: old path doesn't exist
		assertFileNotExistsOnDisk(t, oldPath)
	})

	t.Run("renames folder and updates children paths", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		folder := createTestFolder(t, ctx, fs, userHome, "oldfolder")
		subfolder := createTestFolder(t, ctx, fs, folder, "sub")
		_ = createTestFile(t, ctx, fs, subfolder, "file.txt", []byte("content"))

		err = fs.RenameFile(ctx, folder, "newfolder")
		require.NoError(t, err)

		// Verify: all descendant paths updated
		newFolderPath := file_model.UsersRootPath.Child("testuser/newfolder", true)
		newSubfolderPath := file_model.UsersRootPath.Child("testuser/newfolder/sub", true)
		newFilePath := file_model.UsersRootPath.Child("testuser/newfolder/sub/file.txt", false)

		_, err = fs.GetFileByFilepath(ctx, newFolderPath)
		assert.NoError(t, err)
		assertFileExistsOnDisk(t, newFolderPath)

		_, err = fs.GetFileByFilepath(ctx, newSubfolderPath)
		assert.NoError(t, err)
		assertFileExistsOnDisk(t, newSubfolderPath)

		_, err = fs.GetFileByFilepath(ctx, newFilePath)
		assert.NoError(t, err)
		assertFileExistsOnDisk(t, newFilePath)
	})

	t.Run("fails when name already exists", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file1 := createTestFile(t, ctx, fs, userHome, "file1.txt", []byte("content1"))
		_ = createTestFile(t, ctx, fs, userHome, "file2.txt", []byte("content2"))

		// Try to rename file1 to file2 (should fail)
		err = fs.RenameFile(ctx, file1, "file2.txt")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileAlreadyExists)

		// Verify: no changes
		originalPath := file_model.UsersRootPath.Child("testuser/file1.txt", false)
		assertFileExistsOnDisk(t, originalPath)
	})
}

func TestFileService_DeleteFiles_Integration(t *testing.T) {
	t.Run("deletes file on core tower", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t, withTowerRole(tower_model.RoleCore))
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "delete.txt", []byte("delete me"))
		fileID := file.ID()
		filePath := file.GetPortablePath()

		err = fs.DeleteFiles(ctx, file)
		require.NoError(t, err)

		// Verify: file removed from service
		assertFileNotInService(t, ctx, fs, fileID)

		// Verify: file removed from original location
		assertFileNotExistsOnDisk(t, filePath)
	})

	t.Run("deletes folder recursively", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		folder := createTestFolder(t, ctx, fs, userHome, "todelete")
		subfolder := createTestFolder(t, ctx, fs, folder, "sub")
		file := createTestFile(t, ctx, fs, subfolder, "file.txt", []byte("content"))

		folderID := folder.ID()
		subfolderID := subfolder.ID()
		fileID := file.ID()

		err = fs.DeleteFiles(ctx, folder)
		require.NoError(t, err)

		// Verify: all descendants removed from service
		assertFileNotInService(t, ctx, fs, folderID)
		assertFileNotInService(t, ctx, fs, subfolderID)
		assertFileNotInService(t, ctx, fs, fileID)

		// Verify: folder removed from disk
		folderPath := file_model.UsersRootPath.Child("testuser/todelete", true)
		assertFileNotExistsOnDisk(t, folderPath)
	})

	t.Run("deletes multiple files", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file1 := createTestFile(t, ctx, fs, userHome, "del1.txt", []byte("content1"))
		file2 := createTestFile(t, ctx, fs, userHome, "del2.txt", []byte("content2"))
		file3 := createTestFile(t, ctx, fs, userHome, "del3.txt", []byte("content3"))

		err = fs.DeleteFiles(ctx, file1, file2, file3)
		require.NoError(t, err)

		// Verify: all files removed
		assertFileNotInService(t, ctx, fs, file1.ID())
		assertFileNotInService(t, ctx, fs, file2.ID())
		assertFileNotInService(t, ctx, fs, file3.ID())
	})

	t.Run("prevents deleting user home directory", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		err = fs.DeleteFiles(ctx, userHome)
		assert.Error(t, err)
	})

	t.Run("prevents deleting user trash directory", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		trashPath := file_model.UsersRootPath.Child("testuser/"+file_model.UserTrashDirName, true)
		trash, err := fs.GetFileByFilepath(ctx, trashPath)
		require.NoError(t, err)

		err = fs.DeleteFiles(ctx, trash)
		assert.Error(t, err)
	})
}

func TestFileService_GetChildren_Integration(t *testing.T) {
	t.Run("returns children of folder", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create test folder with children
		folder := createTestFolder(t, ctx, fs, userHome, "parent")
		_ = createTestFile(t, ctx, fs, folder, "child1.txt", []byte("content1"))
		_ = createTestFile(t, ctx, fs, folder, "child2.txt", []byte("content2"))
		_ = createTestFolder(t, ctx, fs, folder, "subfolder")

		children, err := fs.GetChildren(ctx, folder)
		require.NoError(t, err)
		assert.Len(t, children, 3)
	})

	t.Run("returns error for non-directory", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "notfolder.txt", []byte("content"))

		_, err = fs.GetChildren(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrDirectoryRequired)
	})
}

func TestFileService_RecursiveEnsureChildrenLoaded_Integration(t *testing.T) {
	t.Run("loads all children recursively", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create nested structure
		folder := createTestFolder(t, ctx, fs, userHome, "root")
		sub1 := createTestFolder(t, ctx, fs, folder, "sub1")
		sub2 := createTestFolder(t, ctx, fs, folder, "sub2")
		_ = createTestFile(t, ctx, fs, sub1, "file1.txt", []byte("content1"))
		_ = createTestFile(t, ctx, fs, sub2, "file2.txt", []byte("content2"))

		err = fs.RecursiveEnsureChildrenLoaded(ctx, folder)
		require.NoError(t, err)

		// Verify all children are accessible
		assert.True(t, folder.ChildrenLoaded())
		assert.True(t, sub1.ChildrenLoaded())
		assert.True(t, sub2.ChildrenLoaded())
	})

	t.Run("returns error for non-directory", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "notfolder.txt", []byte("content"))

		err = fs.RecursiveEnsureChildrenLoaded(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrDirectoryRequired)
	})
}

func TestFileService_MoveFiles_EmptyList(t *testing.T) {
	t.Run("handles empty file list", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")

		// Move empty list should succeed without error
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{}, destFolder)
		assert.NoError(t, err)
	})
}

func TestFileService_DeleteFiles_EmptyList(t *testing.T) {
	t.Run("handles empty file list", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)

		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Delete empty list should succeed without error
		err := fs.DeleteFiles(ctx)
		assert.NoError(t, err)
	})
}

func TestFileService_GetMediaCacheByFilename(t *testing.T) {
	t.Run("returns error for non-existent cache file", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		_, err := fs.GetMediaCacheByFilename(ctx, "nonexistent.jpg")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNotFound)
	})
}

func TestFileService_AddFile_Integration(t *testing.T) {
	t.Run("returns error for file without ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create a file without ID
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:    userHome.GetPortablePath().Child("noID.txt", false),
			MemOnly: true,
		})

		err = file.SetParent(userHome)
		require.NoError(t, err)

		err = fs.AddFile(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrNoFileID)
	})

	t.Run("returns error for directory file without parent", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Create a directory file without parent (directories don't need content ID)
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("orphan/", true),
			MemOnly:    true,
			GenerateID: true,
		})

		err := fs.AddFile(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrNoParent)
	})
}

func TestFileService_GetFileByID_Integration(t *testing.T) {
	t.Run("returns file by ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "findme.txt", []byte("content"))

		found, err := fs.GetFileByID(ctx, file.ID())
		require.NoError(t, err)
		assert.Equal(t, file.ID(), found.ID())
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		_, err := fs.GetFileByID(ctx, "nonexistent-id")
		assert.Error(t, err)
	})
}

func TestFileService_CreateFile_Integration(t *testing.T) {
	t.Run("creates empty file", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file, err := fs.CreateFile(ctx, userHome, "empty.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		assertFileExistsOnDisk(t, file.GetPortablePath())
	})
}

func TestFileService_CreateFolder_Integration(t *testing.T) {
	t.Run("returns error when folder already exists", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		_, err = fs.CreateFolder(ctx, userHome, "duplicate")
		require.NoError(t, err)

		// Try to create same folder again
		_, err = fs.CreateFolder(ctx, userHome, "duplicate")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrDirectoryAlreadyExists)
	})
}

func TestFileService_ResizeUp_Integration(t *testing.T) {
	t.Run("updates size of parent", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fsSvc := appCtx.GetFileService().(*ServiceImpl)

		userHome, err := fsSvc.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fsSvc, userHome, "sizeme.txt", []byte("some content"))

		err = fsSvc.ResizeUp(ctx, file)
		assert.NoError(t, err)
	})
}

func TestFileService_DeleteFiles_BackupRole(t *testing.T) {
	t.Run("deletes file on backup tower", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t, withTowerRole(tower_model.RoleBackup))
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "backupdelete.txt", []byte("backup content"))
		filePath := file.GetPortablePath()

		err = fs.DeleteFiles(ctx, file)
		require.NoError(t, err)

		// Verify: file removed from original location
		assertFileNotExistsOnDisk(t, filePath)
	})
}

func TestMakeUniqueChildName(t *testing.T) {
	t.Run("returns original name when no conflict", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		uniquePath, err := MakeUniqueChildName(userHome.GetPortablePath(), "newfile.txt", false)
		require.NoError(t, err)
		assert.Equal(t, "newfile.txt", uniquePath.Filename())
	})

	t.Run("appends number when name conflicts", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create a file that will conflict
		_ = createTestFile(t, ctx, fs, userHome, "conflict.txt", []byte("content"))

		uniquePath, err := MakeUniqueChildName(userHome.GetPortablePath(), "conflict.txt", false)
		require.NoError(t, err)
		assert.Equal(t, "conflict.txt (1)", uniquePath.Filename())
	})

	t.Run("handles multiple conflicts", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create files that will conflict
		_ = createTestFile(t, ctx, fs, userHome, "multi.txt", []byte("content"))
		_ = createTestFile(t, ctx, fs, userHome, "multi.txt (1)", []byte("content"))
		_ = createTestFile(t, ctx, fs, userHome, "multi.txt (2)", []byte("content"))

		uniquePath, err := MakeUniqueChildName(userHome.GetPortablePath(), "multi.txt", false)
		require.NoError(t, err)
		assert.Equal(t, "multi.txt (3)", uniquePath.Filename())
	})

	t.Run("handles folder names", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create a folder that will conflict
		_ = createTestFolder(t, ctx, fs, userHome, "myfolder")

		uniquePath, err := MakeUniqueChildName(userHome.GetPortablePath(), "myfolder", true)
		require.NoError(t, err)
		assert.Equal(t, "myfolder (1)", uniquePath.Filename())
	})
}

func TestFileService_MoveFiles_WithNameConflict(t *testing.T) {
	t.Run("renames file when destination has conflict", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create source file
		sourceFile := createTestFile(t, ctx, fs, userHome, "source.txt", []byte("source content"))

		// Create destination folder with a file of the same name
		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")
		_ = createTestFile(t, ctx, fs, destFolder, "source.txt", []byte("existing content"))

		// Move the file - it should get renamed
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{sourceFile}, destFolder)
		require.NoError(t, err)

		// Verify: moved file has a unique name
		expectedPath := file_model.UsersRootPath.Child("testuser/dest/source.txt (1)", false)
		_, err = fs.GetFileByFilepath(ctx, expectedPath)
		assert.NoError(t, err)
		assertFileExistsOnDisk(t, expectedPath)
	})
}

func TestFileService_GetFileByFilepath_Variations(t *testing.T) {
	t.Run("gets nested file by path", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create nested structure
		folder := createTestFolder(t, ctx, fs, userHome, "level1")
		subfolder := createTestFolder(t, ctx, fs, folder, "level2")
		file := createTestFile(t, ctx, fs, subfolder, "deep.txt", []byte("content"))

		// Get by full path
		deepPath := file_model.UsersRootPath.Child("testuser/level1/level2/deep.txt", false)
		found, err := fs.GetFileByFilepath(ctx, deepPath)
		require.NoError(t, err)
		assert.Equal(t, file.ID(), found.ID())
	})

	t.Run("returns error for non-existent path", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		nonExistentPath := file_model.UsersRootPath.Child("testuser/nonexistent/path/file.txt", false)
		_, err := fs.GetFileByFilepath(ctx, nonExistentPath)
		assert.Error(t, err)
	})
}

func TestFileService_RenameFile_SameName(t *testing.T) {
	t.Run("handles rename to same name", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "samename.txt", []byte("content"))

		// Renaming to same name should work (no-op)
		err = fs.RenameFile(ctx, file, "samename.txt")
		assert.Error(t, err) // File already exists with same name
	})
}

func TestFileService_AddFile_ExistingFile(t *testing.T) {
	t.Run("handles adding file that already exists in service", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create a file normally
		file := createTestFile(t, ctx, fs, userHome, "exists.txt", []byte("content"))

		// Try to add the same file again - should be handled gracefully
		err = fs.AddFile(ctx, file)
		assert.NoError(t, err) // Skips duplicate silently
	})
}

func TestFileService_GetChildren_LoadsNewChildren(t *testing.T) {
	t.Run("loads children from disk when not loaded", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create folder
		folder := createTestFolder(t, ctx, fs, userHome, "loadtest")

		// Create files inside the folder
		_ = createTestFile(t, ctx, fs, folder, "child1.txt", []byte("content1"))
		_ = createTestFile(t, ctx, fs, folder, "child2.txt", []byte("content2"))

		// Get children - should include all
		children, err := fs.GetChildren(ctx, folder)
		require.NoError(t, err)
		assert.Len(t, children, 2)
	})
}

func TestFileService_MoveFiles_SkipJournal(t *testing.T) {
	t.Run("moves file with skip journal context", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		sourceFile := createTestFile(t, ctx, fs, userHome, "skipjournal.txt", []byte("content"))
		destFolder := createTestFolder(t, ctx, fs, userHome, "dest")

		// Add skip journal to context
		skipCtx := context.WithValue(ctx, SkipJournalKey, true) //nolint:revive,staticcheck

		err = fs.MoveFiles(skipCtx, []*file_model.WeblensFileImpl{sourceFile}, destFolder)
		require.NoError(t, err)

		// Verify file moved
		expectedPath := file_model.UsersRootPath.Child("testuser/dest/skipjournal.txt", false)
		_, err = fs.GetFileByFilepath(ctx, expectedPath)
		assert.NoError(t, err)
	})
}

func TestFileService_DeleteFiles_SkipJournal(t *testing.T) {
	t.Run("deletes file with skip journal context", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t, withTowerRole(tower_model.RoleBackup))
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "deleteskip.txt", []byte("content"))
		filePath := file.GetPortablePath()

		// Add skip journal to context
		skipCtx := context.WithValue(ctx, SkipJournalKey, true) //nolint:revive,staticcheck

		err = fs.DeleteFiles(skipCtx, file)
		require.NoError(t, err)

		assertFileNotExistsOnDisk(t, filePath)
	})
}

func TestFileService_CreateFolder_Success(t *testing.T) {
	t.Run("creates folder successfully", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		folder, err := fs.CreateFolder(ctx, userHome, "newfolder")
		require.NoError(t, err)
		assert.NotNil(t, folder)
		assert.True(t, folder.IsDir())
		assertFileExistsOnDisk(t, folder.GetPortablePath())
	})
}

func TestFileService_DeleteFiles_MultipleFromSameParent(t *testing.T) {
	t.Run("deletes multiple files from same parent correctly", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create parent folder with multiple children
		parent := createTestFolder(t, ctx, fs, userHome, "parent")
		file1 := createTestFile(t, ctx, fs, parent, "child1.txt", []byte("content1"))
		file2 := createTestFile(t, ctx, fs, parent, "child2.txt", []byte("content2"))
		file3 := createTestFile(t, ctx, fs, parent, "child3.txt", []byte("content3"))

		// Delete all children
		err = fs.DeleteFiles(ctx, file1, file2, file3)
		require.NoError(t, err)

		// Verify parent still exists but is empty
		children, err := fs.GetChildren(ctx, parent)
		require.NoError(t, err)
		assert.Empty(t, children)
	})
}

func TestFileService_MoveFiles_FromMultipleParents(t *testing.T) {
	t.Run("moves files from different parents to same destination", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create two source folders
		source1 := createTestFolder(t, ctx, fs, userHome, "source1")
		source2 := createTestFolder(t, ctx, fs, userHome, "source2")
		dest := createTestFolder(t, ctx, fs, userHome, "dest")

		// Create files in different parents
		file1 := createTestFile(t, ctx, fs, source1, "file1.txt", []byte("content1"))
		file2 := createTestFile(t, ctx, fs, source2, "file2.txt", []byte("content2"))

		// Move both to destination
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{file1, file2}, dest)
		require.NoError(t, err)

		// Verify both files moved
		path1 := file_model.UsersRootPath.Child("testuser/dest/file1.txt", false)
		path2 := file_model.UsersRootPath.Child("testuser/dest/file2.txt", false)

		assertFileExistsOnDisk(t, path1)
		assertFileExistsOnDisk(t, path2)
	})
}

func TestFileService_GetFileByFilepath_DontLoad(t *testing.T) {
	t.Run("returns error when dontLoadNew is true and path not in service", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Try to get a path with dontLoadNew flag
		nonExistentPath := file_model.UsersRootPath.Child("testuser/notloaded/deep.txt", false)
		_, err := fs.GetFileByFilepath(ctx, nonExistentPath, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNotFound)
	})
}

func TestFileService_RenameFile_WithNestedChildren(t *testing.T) {
	t.Run("renames deeply nested folder structure", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create deep structure: root/level1/level2/level3/file.txt
		root := createTestFolder(t, ctx, fs, userHome, "root")
		level1 := createTestFolder(t, ctx, fs, root, "level1")
		level2 := createTestFolder(t, ctx, fs, level1, "level2")
		level3 := createTestFolder(t, ctx, fs, level2, "level3")
		_ = createTestFile(t, ctx, fs, level3, "deepfile.txt", []byte("deep content"))

		// Rename the top folder
		err = fs.RenameFile(ctx, root, "renamed")
		require.NoError(t, err)

		// Verify all nested paths are updated
		deepFilePath := file_model.UsersRootPath.Child("testuser/renamed/level1/level2/level3/deepfile.txt", false)
		_, err = fs.GetFileByFilepath(ctx, deepFilePath)
		assert.NoError(t, err)
		assertFileExistsOnDisk(t, deepFilePath)
	})
}

func TestFileService_DeleteFiles_WithNestedStructure(t *testing.T) {
	t.Run("deletes deeply nested folder structure", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create deep structure with multiple files at each level
		root := createTestFolder(t, ctx, fs, userHome, "deeproot")
		_ = createTestFile(t, ctx, fs, root, "file1.txt", []byte("content1"))
		level1 := createTestFolder(t, ctx, fs, root, "level1")
		_ = createTestFile(t, ctx, fs, level1, "file2.txt", []byte("content2"))
		level2 := createTestFolder(t, ctx, fs, level1, "level2")
		_ = createTestFile(t, ctx, fs, level2, "file3.txt", []byte("content3"))

		rootPath := root.GetPortablePath()

		err = fs.DeleteFiles(ctx, root)
		require.NoError(t, err)

		// Verify entire structure is gone
		assertFileNotExistsOnDisk(t, rootPath)
	})
}

func TestFileService_MoveFiles_NestedFolder(t *testing.T) {
	t.Run("moves nested folder preserving structure", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create source with nested structure
		source := createTestFolder(t, ctx, fs, userHome, "source")
		nested := createTestFolder(t, ctx, fs, source, "nested")
		deepNested := createTestFolder(t, ctx, fs, nested, "deepnested")
		file := createTestFile(t, ctx, fs, deepNested, "file.txt", []byte("content"))
		fileID := file.ID()

		// Create destination
		dest := createTestFolder(t, ctx, fs, userHome, "dest")

		// Move the source folder
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{source}, dest)
		require.NoError(t, err)

		// Verify deep file is accessible at new path
		newDeepPath := file_model.UsersRootPath.Child("testuser/dest/source/nested/deepnested/file.txt", false)
		movedFile, err := fs.GetFileByFilepath(ctx, newDeepPath)
		require.NoError(t, err)
		assert.Equal(t, fileID, movedFile.ID())
		assertFileExistsOnDisk(t, newDeepPath)

		// Verify old path doesn't exist
		oldPath := file_model.UsersRootPath.Child("testuser/source", true)
		assertFileNotExistsOnDisk(t, oldPath)
	})
}

func TestFileService_CreateFolder_NestedInNewFolder(t *testing.T) {
	t.Run("creates nested folder in newly created folder", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create first folder
		folder1, err := fs.CreateFolder(ctx, userHome, "first")
		require.NoError(t, err)

		// Immediately create nested folder
		folder2, err := fs.CreateFolder(ctx, folder1, "second")
		require.NoError(t, err)

		// And another level
		folder3, err := fs.CreateFolder(ctx, folder2, "third")
		require.NoError(t, err)

		assert.True(t, folder3.IsDir())
		assertFileExistsOnDisk(t, folder3.GetPortablePath())
	})
}

func TestFileService_CreateFile_InNewFolder(t *testing.T) {
	t.Run("creates file in newly created folder", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create folder
		folder, err := fs.CreateFolder(ctx, userHome, "newfolder")
		require.NoError(t, err)

		// Immediately create file in it
		file, err := fs.CreateFile(ctx, folder, "newfile.txt")
		require.NoError(t, err)

		assert.NotNil(t, file)
		assertFileExistsOnDisk(t, file.GetPortablePath())
	})
}

func TestFileService_GetFileByID_RootFile(t *testing.T) {
	t.Run("gets USERS root by ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Get root file by tree key
		root, err := fs.GetFileByID(ctx, file_model.UsersTreeKey)
		require.NoError(t, err)
		assert.NotNil(t, root)
		assert.True(t, root.IsDir())
	})

	t.Run("gets RESTORE root by ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		root, err := fs.GetFileByID(ctx, file_model.RestoreTreeKey)
		require.NoError(t, err)
		assert.NotNil(t, root)
	})

	t.Run("gets BACKUP root by ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		root, err := fs.GetFileByID(ctx, file_model.BackupTreeKey)
		require.NoError(t, err)
		assert.NotNil(t, root)
	})
}

func TestFileService_RecursiveEnsureChildrenLoaded_NonDirectory(t *testing.T) {
	t.Run("returns error for non-directory", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "notfolder.txt", []byte("content"))

		err = fs.RecursiveEnsureChildrenLoaded(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrDirectoryRequired)
	})
}

func TestFileService_GetFileByID_NotFound(t *testing.T) {
	t.Run("returns error for non-existent ID", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		_, err := fs.GetFileByID(ctx, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNotFound)
	})
}

func TestFileService_GetFileByFilepath_NotFound(t *testing.T) {
	t.Run("returns error for non-existent path", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		_, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("nonexistent/path.txt", false))
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNotFound)
	})
}

func TestFileService_GetChildren_EmptyFolder(t *testing.T) {
	t.Run("returns empty slice for empty folder", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		emptyFolder := createTestFolder(t, ctx, fs, userHome, "emptyfolder")

		children, err := fs.GetChildren(ctx, emptyFolder)
		require.NoError(t, err)
		assert.Len(t, children, 0)
	})
}

func TestFileService_AddFile_NilParent(t *testing.T) {
	t.Run("returns error when parent not in service", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		// Create a directory with parent that's not in the file service
		// Use directory path (ends with /) to avoid contentID check
		file := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       file_model.UsersRootPath.Child("nonexistent_parent/subdir/", true),
			MemOnly:    true,
			GenerateID: true,
		})

		err := fs.AddFile(ctx, file)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrNoParent)
	})
}

func TestFileService_MoveFiles_SameLocation(t *testing.T) {
	t.Run("handles move to same parent", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "stayhome.txt", []byte("content"))

		// Moving to the same folder should work (no-op or rename)
		err = fs.MoveFiles(ctx, []*file_model.WeblensFileImpl{file}, userHome)
		require.NoError(t, err)
	})
}

func TestFileService_RenameFile_EmptyName(t *testing.T) {
	t.Run("returns error for empty name", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		file := createTestFile(t, ctx, fs, userHome, "original.txt", []byte("content"))

		err = fs.RenameFile(ctx, file, "")
		assert.Error(t, err)
	})
}

func TestFileService_DeleteFiles_Recursive(t *testing.T) {
	t.Run("deletes folder with deep nesting", func(t *testing.T) {
		ctx, _ := newIntegrationTestContext(t)
		appCtx, ok := ctxservice.FromContext(ctx)
		require.True(t, ok)

		fs := appCtx.GetFileService()

		userHome, err := fs.GetFileByFilepath(ctx, file_model.UsersRootPath.Child("testuser", true))
		require.NoError(t, err)

		// Create deeply nested structure
		level1 := createTestFolder(t, ctx, fs, userHome, "level1")
		level2 := createTestFolder(t, ctx, fs, level1, "level2")
		level3 := createTestFolder(t, ctx, fs, level2, "level3")
		_ = createTestFile(t, ctx, fs, level3, "deep.txt", []byte("deep content"))
		_ = createTestFile(t, ctx, fs, level2, "mid.txt", []byte("mid content"))
		_ = createTestFile(t, ctx, fs, level1, "top.txt", []byte("top content"))

		// Delete the top-level folder
		err = fs.DeleteFiles(ctx, level1)
		require.NoError(t, err)

		// Verify folder is gone
		_, err = fs.GetFileByID(ctx, level1.ID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, file_model.ErrFileNotFound)
	})
}
