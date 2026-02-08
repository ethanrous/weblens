package e2e_test

import (
	"net/http"
	"testing"

	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFolder(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()
	require.NotEmpty(t, homeID)

	// Scan the home folder
	taskInfo, resp, err := client.FolderAPI.ScanFolder(t.Context(), homeID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, taskInfo.GetTaskID())
}

func TestGetFolderHistory(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()
	require.NotEmpty(t, homeID)

	// Get folder history - returns action history for the folder
	history, resp, err := client.FolderAPI.GetFolderHistory(t.Context(), homeID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// History array should be non-nil (may have entries from folder creation)
	assert.NotNil(t, history)
}
