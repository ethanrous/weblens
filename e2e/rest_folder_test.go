package e2e_test

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/auth"
	media_model "github.com/ethanrous/weblens/models/media"
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

// TestSetFolderCover_RejectsForeignMedia verifies that a user with edit
// permission on their own folder cannot set its cover to a contentID whose
// only backing files belong to another user. Previously this leaked the
// media's owner, FileIDs, and Location through the response notification.
func TestSetFolderCover_RejectsForeignMedia(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	adminFolder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: adminUser.GetHomeID(),
		NewFolderName:  "admin-private",
	}).Execute()
	require.NoError(t, err)

	fileID, victimContentID := uploadTestImage(t, coreSetup, adminClient, adminFolder.GetId())

	syntheticMedia := &media_model.Media{
		ContentID:  media_model.ContentID(victimContentID),
		Owner:      "admin",
		MimeType:   "image/jpeg",
		FileIDs:    []string{fileID},
		Width:      1920,
		Height:     1080,
		PageCount:  1,
		CreateDate: time.Unix(1000, 0),
		Enabled:    true,
	}
	err = media_model.SaveMedia(coreSetup.ctx, syntheticMedia)
	require.NoError(t, err)

	autoActivate := true
	_, err = adminClient.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "coverthief",
		Password:     "TestPass123",
		FullName:     "Cover Thief",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	attackerToken, err := auth.GenerateNewToken(coreSetup.ctx, "coverthief-token", "coverthief", coreSetup.ctx.LocalTowerID)
	require.NoError(t, err)

	attackerClient := getAPIClientFromConfig(coreSetup.cnf, base64.StdEncoding.EncodeToString(attackerToken.Token[:]))

	attackerUser, _, err := attackerClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	attackerFolder, _, err := attackerClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: attackerUser.GetHomeID(),
		NewFolderName:  "attacker-folder",
	}).Execute()
	require.NoError(t, err)

	resp, err := attackerClient.FolderAPI.SetFolderCover(t.Context(), attackerFolder.GetId()).
		ContentID(victimContentID).
		Execute()
	assert.Error(t, err, "attacker must not be allowed to set cover to a media they cannot view")
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
