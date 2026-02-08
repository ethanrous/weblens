package e2e_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCoreAndBackup(t *testing.T) (setupResult, setupResult) {
	// Setup a core server for the backup to connect to
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start core test server")

	// Setup a backup server
	backupSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleBackup), CoreAddress: coreSetup.address, CoreToken: coreSetup.token})
	require.NoError(t, err, "Failed to start backup test server")

	return coreSetup, backupSetup
}

func launchBackupAndWait(t *testing.T, backupSetup setupResult) *task.Task {
	for len(backupSetup.ctx.ClientService.GetAllClients()) == 0 {
		log.GlobalLogger().Debug().Msg("Test is waiting for backup to connect to core...")
		time.Sleep(10 * time.Millisecond)
	}

	coreInstance := backupSetup.ctx.ClientService.GetAllClients()[0].GetInstance()
	tsk, err := jobs.BackupOne(backupSetup.ctx, *coreInstance)
	assert.NoError(t, err)

	tsk.Wait()
	err = tsk.ReadError()
	require.NoError(t, err)

	complete, result := tsk.Status()
	require.True(t, complete)
	require.Equal(t, task.TaskSuccess, result)

	return tsk
}

func TestBackupFiles(t *testing.T) {
	coreSetup, backupSetup := newCoreAndBackup(t)

	// Add a file to the core server
	adminHomeDir, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Child("admin", true))
	require.NoError(t, err)

	fileContent := []byte("This is a test file.")
	_, err = coreSetup.ctx.FileService.CreateFile(coreSetup.ctx, adminHomeDir, "test-file", fileContent)
	require.NoError(t, err)

	// Launch the backup task on the backup server
	launchBackupAndWait(t, backupSetup)

	// Verify that the file was copied to the backup server
	createdFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Children(coreSetup.ctx.GetTowerID(), "admin", "test-file").AsFile())
	if assert.NoError(t, err) {
		assert.True(t, createdFile.Exists())

		copiedFileContent, err := createdFile.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, string(fileContent), string(copiedFileContent))
	}
}

func TestBackupFilesInDirectories(t *testing.T) {
	coreSetup, backupSetup := newCoreAndBackup(t)

	// Add a file to the core server
	adminHomeDir, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Child("admin", true))
	require.NoError(t, err)

	folder, err := coreSetup.ctx.FileService.CreateFolder(coreSetup.ctx, adminHomeDir, "test-folder")
	require.NoError(t, err)

	fileContent := []byte("This is a test file in a folder.")
	_, err = coreSetup.ctx.FileService.CreateFile(coreSetup.ctx, folder, "test-file", fileContent)
	require.NoError(t, err)

	// Launch the backup task on the backup server
	launchBackupAndWait(t, backupSetup)

	// Verify that the file was copied to the backup server
	createdFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Children(coreSetup.ctx.GetTowerID(), "admin", "test-folder", "test-file").AsFile())
	if assert.NoError(t, err) {
		assert.True(t, createdFile.Exists())

		copiedFileContent, err := createdFile.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, string(fileContent), string(copiedFileContent))
	}
}

func TestBackupFilesMoved(t *testing.T) {
	coreSetup, backupSetup := newCoreAndBackup(t)

	// Add a file to the core server
	adminHomeDir, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Child("admin", true))
	require.NoError(t, err)

	folder, err := coreSetup.ctx.FileService.CreateFolder(coreSetup.ctx, adminHomeDir, "test-folder")
	require.NoError(t, err)

	fileContent := []byte("This is a test file in a folder.")
	_, err = coreSetup.ctx.FileService.CreateFile(coreSetup.ctx, folder, "test-file", fileContent)
	require.NoError(t, err)

	// Launch the backup task on the backup server
	launchBackupAndWait(t, backupSetup)

	// Move the file to a new location on the core server
	newFolder, err := coreSetup.ctx.FileService.CreateFolder(coreSetup.ctx, adminHomeDir, "new-folder")
	require.NoError(t, err)

	// Get the original file
	originalFile, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Children("admin", "test-folder", "test-file").AsFile())
	require.NoError(t, err)

	// Move the file to the new folder
	err = coreSetup.ctx.FileService.MoveFiles(coreSetup.ctx, []*file.WeblensFileImpl{originalFile}, newFolder)
	require.NoError(t, err)

	// Launch the backup task again on the backup server
	launchBackupAndWait(t, backupSetup)

	// Verify that the file was copied to the backup server
	createdFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Children(coreSetup.ctx.GetTowerID(), "admin", "new-folder", "test-file").AsFile())
	if assert.NoError(t, err) {
		assert.True(t, createdFile.Exists())

		copiedFileContent, err := createdFile.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, string(fileContent), string(copiedFileContent))
	}
}

func TestBackupFilesDeleted(t *testing.T) {
	coreSetup, backupSetup := newCoreAndBackup(t)

	// Add a file to the core server
	adminHomeDir, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Child("admin", true))
	require.NoError(t, err)

	fileContent := []byte("This is a test file, soon to be deleted.")
	_, err = coreSetup.ctx.FileService.CreateFile(coreSetup.ctx, adminHomeDir, "test-file", fileContent)
	require.NoError(t, err)

	// Launch the backup task on the backup server
	launchBackupAndWait(t, backupSetup)

	// Verify that the file was copied to the backup server
	createdFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Children(coreSetup.ctx.GetTowerID(), "admin", "test-file").AsFile())
	require.NoError(t, err)
	require.True(t, createdFile.Exists())

	// Delete the file from the core server
	originalFile, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Children("admin", "test-file").AsFile())
	require.NoError(t, err)

	err = coreSetup.ctx.FileService.DeleteFiles(coreSetup.ctx, originalFile)
	require.NoError(t, err)

	// Launch the backup task again on the backup server
	launchBackupAndWait(t, backupSetup)

	// Verify that the file is no longer present on the backup server
	deletedFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Children(coreSetup.ctx.GetTowerID(), "admin", "test-file").AsFile())
	if assert.NoError(t, err) {
		assert.False(t, deletedFile.Exists())
	}
}
