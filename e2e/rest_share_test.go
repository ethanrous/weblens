package e2e_test

import (
	"net/http"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFileShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get user's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "shared-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share for the folder
	shareInfo, resp, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
		Public: openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, shareInfo.GetShareID())
	assert.Equal(t, folder.GetId(), shareInfo.GetFileID())
	assert.False(t, shareInfo.GetPublic())
}

func TestGetFileShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get user's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "get-share-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
		Public: openapi.PtrBool(true),
	}).Execute()
	require.NoError(t, err)

	// Get the share
	shareInfo, resp, err := client.ShareAPI.GetFileShare(t.Context(), createdShare.GetShareID()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, createdShare.GetShareID(), shareInfo.GetShareID())
	assert.Equal(t, folder.GetId(), shareInfo.GetFileID())
	assert.True(t, shareInfo.GetPublic())

	// Test getting non-existent share
	_, resp, err = client.ShareAPI.GetFileShare(t.Context(), "non-existent-share").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetSharePublic(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get user's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "public-toggle-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a non-public share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
		Public: openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)
	assert.False(t, createdShare.GetPublic())

	// Set share to public
	resp, err := client.ShareAPI.SetSharePublic(t.Context(), createdShare.GetShareID()).Public(true).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the change
	shareInfo, _, err := client.ShareAPI.GetFileShare(t.Context(), createdShare.GetShareID()).Execute()
	require.NoError(t, err)
	assert.True(t, shareInfo.GetPublic())

	// Set share back to private
	resp, err = client.ShareAPI.SetSharePublic(t.Context(), createdShare.GetShareID()).Public(false).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the change
	shareInfo, _, err = client.ShareAPI.GetFileShare(t.Context(), createdShare.GetShareID()).Execute()
	require.NoError(t, err)
	assert.False(t, shareInfo.GetPublic())
}

func TestAddUserToShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create another user to add to the share
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "shareuser",
		Password:     "TestPass123",
		FullName:     "Share User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Get admin's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "add-user-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
	}).Execute()
	require.NoError(t, err)

	// Add user to share
	shareInfo, resp, err := client.ShareAPI.AddUserToShare(t.Context(), createdShare.GetShareID()).Request(openapi.AddUserParams{
		Username:    "shareuser",
		CanView:     openapi.PtrBool(true),
		CanEdit:     openapi.PtrBool(false),
		CanDownload: openapi.PtrBool(true),
		CanDelete:   openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify user was added
	found := false

	for _, accessor := range shareInfo.GetAccessors() {
		if accessor.GetUsername() == "shareuser" {
			found = true

			break
		}
	}

	assert.True(t, found, "shareuser should be in accessors list")
}

func TestRemoveUserFromShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create another user to add then remove from the share
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "removableuser",
		Password:     "TestPass123",
		FullName:     "Removable User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Get admin's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "remove-user-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
	}).Execute()
	require.NoError(t, err)

	// Add user to share
	_, _, err = client.ShareAPI.AddUserToShare(t.Context(), createdShare.GetShareID()).Request(openapi.AddUserParams{
		Username: "removableuser",
		CanView:  openapi.PtrBool(true),
	}).Execute()
	require.NoError(t, err)

	// Remove user from share
	shareInfo, resp, err := client.ShareAPI.RemoveUserFromShare(t.Context(), createdShare.GetShareID(), "removableuser").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify user was removed
	found := false

	for _, accessor := range shareInfo.GetAccessors() {
		if accessor.GetUsername() == "removableuser" {
			found = true

			break
		}
	}

	assert.False(t, found, "removableuser should not be in accessors list")
}

func TestUpdateSharePermissions(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create another user to add to the share
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "permuser",
		Password:     "TestPass123",
		FullName:     "Permissions User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Get admin's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "permissions-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
	}).Execute()
	require.NoError(t, err)

	// Add user to share with view-only permissions
	_, _, err = client.ShareAPI.AddUserToShare(t.Context(), createdShare.GetShareID()).Request(openapi.AddUserParams{
		Username:    "permuser",
		CanView:     openapi.PtrBool(true),
		CanEdit:     openapi.PtrBool(false),
		CanDownload: openapi.PtrBool(false),
		CanDelete:   openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)

	// Update permissions to allow editing and downloading
	shareInfo, resp, err := client.ShareAPI.UpdateShareAccessorPermissions(t.Context(), createdShare.GetShareID(), "permuser").Request(openapi.PermissionsParams{
		CanView:     openapi.PtrBool(true),
		CanEdit:     openapi.PtrBool(true),
		CanDownload: openapi.PtrBool(true),
		CanDelete:   openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify permissions were updated
	permissions := shareInfo.GetPermissions()
	if perms, ok := permissions["permuser"]; ok {
		assert.True(t, perms.GetCanView())
		assert.True(t, perms.GetCanEdit())
		assert.True(t, perms.GetCanDownload())
		assert.False(t, perms.GetCanDelete())
	} else {
		t.Error("permuser permissions not found in share info")
	}
}

func TestDeleteFileShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	if err != nil {
		log.GlobalLogger().Error().Stack().Err(err).Msg("Failed to start test server")
		t.FailNow()
	}

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get user's home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()

	// Create a folder to share
	folder, _, err := client.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "delete-share-folder",
	}).Execute()
	require.NoError(t, err)

	// Create a share
	createdShare, _, err := client.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
	}).Execute()
	require.NoError(t, err)

	shareID := createdShare.GetShareID()

	// Verify share exists
	_, _, err = client.ShareAPI.GetFileShare(t.Context(), shareID).Execute()
	require.NoError(t, err)

	// Delete the share
	resp, err := client.ShareAPI.DeleteFileShare(t.Context(), shareID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify share no longer exists
	_, resp, err = client.ShareAPI.GetFileShare(t.Context(), shareID).Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
