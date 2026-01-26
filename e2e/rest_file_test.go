package e2e_test

import (
	"net/http"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFolder(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()
	require.NotEmpty(t, homeID)

	// Get the home folder
	folderInfo, resp, err := client.FolderAPI.GetFolder(t.Context(), homeID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	self := folderInfo.GetSelf()
	assert.Equal(t, homeID, self.GetId())
	assert.True(t, self.GetIsDir())
}

func TestCreateFolder(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a new folder
	fileInfo, resp, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "test-folder",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	path, err := fs.ParsePortable(fileInfo.GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "test-folder", path.Filename())
	assert.True(t, fileInfo.GetIsDir())

	// Verify folder exists by getting it
	folderInfo, _, err := client.FolderAPI.GetFolder(t.Context(), fileInfo.GetId()).Execute()
	require.NoError(t, err)

	selfInfo := folderInfo.GetSelf()
	assert.Equal(t, fileInfo.GetId(), selfInfo.GetId())
}

func TestGetFile(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to test GetFile on
	createdFolder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "getfile-test",
	}).Execute()
	require.NoError(t, err)

	// Get the file info
	fileInfo, resp, err := client.FilesAPI.GetFile(t.Context(), createdFolder.GetId()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, createdFolder.GetId(), fileInfo.GetId())

	path, err := fs.ParsePortable(fileInfo.GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "getfile-test", path.Filename())

	// Test getting non-existent file
	_, resp, err = client.FilesAPI.GetFile(t.Context(), "non-existent-id").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSearchByFilename(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create folders with different names
	_, _, err = client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "searchable-folder1",
	}).Execute()
	require.NoError(t, err)

	_, _, err = client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "searchable-folder2",
	}).Execute()
	require.NoError(t, err)

	_, _, err = client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "other-folder",
	}).Execute()
	require.NoError(t, err)

	// Search for "searchable" - should find both searchable folders
	results, resp, err := client.FilesAPI.SearchByFilename(t.Context()).Search("searchable").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, len(results))

	// Search with empty query - should fail
	_, resp, err = client.FilesAPI.SearchByFilename(t.Context()).Search("").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateFile(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to rename
	createdFolder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "original-name",
	}).Execute()
	require.NoError(t, err)

	// Rename the folder
	newName := "renamed-folder"
	resp, err := client.FilesAPI.UpdateFile(t.Context(), createdFolder.GetId()).Request(openapi.UpdateFileParams{
		NewName: &newName,
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the name changed
	fileInfo, _, err := client.FilesAPI.GetFile(t.Context(), createdFolder.GetId()).Execute()
	require.NoError(t, err)

	path, err := fs.ParsePortable(fileInfo.GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "renamed-folder", path.Filename())
}

func TestMoveFiles(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create two folders - one will be moved into the other
	targetFolder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "target-folder",
	}).Execute()
	require.NoError(t, err)

	folderToMove, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "folder-to-move",
	}).Execute()
	require.NoError(t, err)

	// Move folder into target folder
	targetID := targetFolder.GetId()
	resp, err := client.FilesAPI.MoveFiles(t.Context()).Request(openapi.MoveFilesParams{
		FileIDs:     []string{folderToMove.GetId()},
		NewParentID: &targetID,
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify folder was moved by checking target folder's children
	targetFolderInfo, _, err := client.FolderAPI.GetFolder(t.Context(), targetFolder.GetId()).Execute()
	require.NoError(t, err)
	assert.Equal(t, 1, len(targetFolderInfo.GetChildren()))

	path, err := fs.ParsePortable(targetFolderInfo.GetChildren()[0].GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "folder-to-move", path.Filename())
}

func TestDeleteFiles(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID and trash ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()
	trashID := userInfo.GetTrashID()

	// Create a folder to delete
	createdFolder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "folder-to-delete",
	}).Execute()
	require.NoError(t, err)

	// First move to trash
	resp, err := client.FilesAPI.MoveFiles(t.Context()).Request(openapi.MoveFilesParams{
		FileIDs:     []string{createdFolder.GetId()},
		NewParentID: &trashID,
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Then delete permanently
	resp, err = client.FilesAPI.DeleteFiles(t.Context()).Request(openapi.FilesListParams{
		FileIDs: []string{createdFolder.GetId()},
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify folder no longer exists
	_, resp, err = client.FilesAPI.GetFile(t.Context(), createdFolder.GetId()).Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
