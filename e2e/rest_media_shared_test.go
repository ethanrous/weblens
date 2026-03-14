package e2e_test

import (
	"encoding/base64"
	"net/http"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMediaBatch_SharedFolder(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err)

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	autoActivate := true
	_, err = adminClient.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "shareviewer",
		Password:     "TestPass123",
		FullName:     "Share Viewer",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	viewerToken, err := auth.GenerateNewToken(coreSetup.ctx, "test-viewer-token", "shareviewer", coreSetup.ctx.LocalTowerID)
	require.NoError(t, err)

	viewerClient := getAPIClientFromConfig(coreSetup.cnf, base64.StdEncoding.EncodeToString(viewerToken.Token[:]))

	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	folder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: adminUser.GetHomeID(),
		NewFolderName:  "shared-photos",
	}).Execute()
	require.NoError(t, err)

	createdShare, _, err := adminClient.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
		Public: openapi.PtrBool(false),
		Users:  []string{"shareviewer"},
	}).Execute()
	require.NoError(t, err)

	shareID := createdShare.GetShareID()
	require.NotEmpty(t, shareID)

	mediaBatch, resp, err := viewerClient.MediaAPI.GetMedia(t.Context()).
		ShareID(shareID).
		FolderIDs([]string{folder.GetId()}).
		Page(0).
		Limit(200).
		Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(0), mediaBatch.GetTotalMediaCount())
}

func TestGetMediaBatch_PublicShare_Unauthenticated(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err)

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	folder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: adminUser.GetHomeID(),
		NewFolderName:  "public-timeline-folder",
	}).Execute()
	require.NoError(t, err)

	createdShare, _, err := adminClient.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()),
		Public: openapi.PtrBool(true),
	}).Execute()
	require.NoError(t, err)

	shareID := createdShare.GetShareID()
	require.NotEmpty(t, shareID)

	unauthClient := getAPIClientFromConfig(coreSetup.cnf, "")

	mediaBatch, resp, err := unauthClient.MediaAPI.GetMedia(t.Context()).
		ShareID(shareID).
		FolderIDs([]string{folder.GetId()}).
		Page(0).
		Limit(200).
		Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(0), mediaBatch.GetTotalMediaCount())
}

func TestGetMediaBatch_SharedFolder_NoShare(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err)

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	autoActivate := true
	_, err = adminClient.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "noShareUser",
		Password:     "TestPass123",
		FullName:     "No Share User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	userToken, err := auth.GenerateNewToken(coreSetup.ctx, "test-token", "noShareUser", coreSetup.ctx.LocalTowerID)
	require.NoError(t, err)

	userClient := getAPIClientFromConfig(coreSetup.cnf, base64.StdEncoding.EncodeToString(userToken.Token[:]))

	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	folder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: adminUser.GetHomeID(),
		NewFolderName:  "private-photos",
	}).Execute()
	require.NoError(t, err)

	_, resp, err := userClient.MediaAPI.GetMedia(t.Context()).
		FolderIDs([]string{folder.GetId()}).
		Page(0).
		Limit(200).
		Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
