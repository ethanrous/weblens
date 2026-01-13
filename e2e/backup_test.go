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
)

func TestBackupFiles(t *testing.T) {
	// Setup a core server for the backup to connect to
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	// Add a file to the core server
	adminHomeDir, err := coreSetup.ctx.FileService.GetFileByFilepath(coreSetup.ctx, file.UsersRootPath.Child("admin", true))
	assert.NoError(t, err)

	fileContent := []byte("This is a test file.")
	_, err = coreSetup.ctx.FileService.CreateFile(coreSetup.ctx, adminHomeDir, "test-file", fileContent)
	assert.NoError(t, err)

	// Setup a backup server
	backupSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleBackup), CoreAddress: coreSetup.address, CoreToken: coreSetup.token})
	if err != nil {
		log.GlobalLogger().Error().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	for len(backupSetup.ctx.ClientService.GetAllClients()) == 0 {
		log.GlobalLogger().Debug().Msg("NO CLIENTS YET!")
		time.Sleep(10 * time.Millisecond)
	}

	coreInstance := backupSetup.ctx.ClientService.GetAllClients()[0].GetInstance()
	tsk, err := jobs.BackupOne(backupSetup.ctx, *coreInstance)
	assert.NoError(t, err)

	tsk.Wait()
	err = tsk.ReadError()
	assert.NoError(t, err)

	complete, result := tsk.Status()
	assert.True(t, complete)
	assert.Equal(t, task.TaskSuccess, result)

	createdFile, err := backupSetup.ctx.FileService.GetFileByFilepath(backupSetup.ctx, file.BackupRootPath.Child(coreSetup.ctx.GetTowerID(), true).Child("admin", true).Child("test-file", false))
	if assert.NoError(t, err) {
		assert.True(t, createdFile.Exists())

		copiedFileContent, err := createdFile.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, string(fileContent), string(copiedFileContent))
	}
}
