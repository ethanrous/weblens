package e2e_test

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZipDownload_CascadesSourceFilePermissions verifies that downloading a
// cached zip re-validates the requester's access to every underlying source
// file. Revoking a share's download permission must invalidate previously
// generated zips.
func TestZipDownload_CascadesSourceFilePermissions(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err)

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	autoActivate := true
	_, err = adminClient.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username: "zipper", Password: "TestPass123", FullName: "Zipper", AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	zipperToken, err := auth.GenerateNewToken(coreSetup.ctx, "zipper-token", "zipper", coreSetup.ctx.LocalTowerID)
	require.NoError(t, err)
	zipperClient := getAPIClientFromConfig(coreSetup.cnf, base64.StdEncoding.EncodeToString(zipperToken.Token[:]))

	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)
	folder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: adminUser.GetHomeID(), NewFolderName: "zip-cascade-folder",
	}).Execute()
	require.NoError(t, err)

	_ = uploadTextFile(t, coreSetup, adminClient, folder.GetId(), "a.txt", "alpha")
	_ = uploadTextFile(t, coreSetup, adminClient, folder.GetId(), "b.txt", "beta")

	share, _, err := adminClient.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folder.GetId()), Users: []string{"zipper"},
	}).Execute()
	require.NoError(t, err)
	shareID := share.GetShareID()

	_, _, err = adminClient.ShareAPI.UpdateShareAccessorPermissions(t.Context(), shareID, "zipper").Request(openapi.PermissionsParams{
		CanView: openapi.PtrBool(true), CanDownload: openapi.PtrBool(true),
	}).Execute()
	require.NoError(t, err)

	// Poll until the zip task completes: 202 means still pending, 200 means cached and ready.
	var takeoutID string
	require.Eventually(t, func() bool {
		info, _, err := zipperClient.FilesAPI.CreateTakeout(t.Context()).ShareID(shareID).Request(openapi.FilesListParams{
			FileIDs: []string{folder.GetId()},
		}).Execute()
		if err != nil {
			return false
		}
		id := info.GetTakeoutID()
		if id != "" {
			takeoutID = id
			return true
		}
		return false
	}, 30*time.Second, 500*time.Millisecond, "zip task should complete and return a TakeoutID")

	require.NotEmpty(t, takeoutID, "TakeoutID must be set after zip task completes")

	// Zipper can download the cached zip while download permission is granted.
	_, resp, err := zipperClient.FilesAPI.DownloadFile(t.Context(), takeoutID).ShareID(shareID).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Revoke download permission; view remains so the user can still see the share.
	_, _, err = adminClient.ShareAPI.UpdateShareAccessorPermissions(t.Context(), shareID, "zipper").Request(openapi.PermissionsParams{
		CanView: openapi.PtrBool(true), CanDownload: openapi.PtrBool(false),
	}).Execute()
	require.NoError(t, err)

	// The cascade re-validates source file permissions and must reject the cached zip.
	_, resp, err = zipperClient.FilesAPI.DownloadFile(t.Context(), takeoutID).ShareID(shareID).Execute()
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"after download perm is revoked, cached zip download must be rejected by source-file cascade")
}
