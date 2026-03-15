package e2e_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

// TestGetMediaBatch_SharedFolder_WithSyntheticMedia verifies that a share recipient
// can query media in a shared folder. Uses direct media insertion to avoid
// depending on image processing infrastructure.
func TestGetMediaBatch_SharedFolder_WithSyntheticMedia(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err)

	adminClient := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create the share recipient user
	autoActivate := true
	_, err = adminClient.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "timelineviewer",
		Password:     "TestPass123",
		FullName:     "Timeline Viewer",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	viewerToken, err := auth.GenerateNewToken(coreSetup.ctx, "test-viewer-token", "timelineviewer", coreSetup.ctx.LocalTowerID)
	require.NoError(t, err)

	viewerClient := getAPIClientFromConfig(coreSetup.cnf, base64.StdEncoding.EncodeToString(viewerToken.Token[:]))

	// Get admin user info
	adminUser, _, err := adminClient.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := adminUser.GetHomeID()
	require.NotEmpty(t, homeID)

	// Create a folder
	folder, _, err := adminClient.FolderAPI.CreateFolder(t.Context()).Request(openapi.CreateFolderBody{
		ParentFolderID: homeID,
		NewFolderName:  "synthetic-media-folder",
	}).Execute()
	require.NoError(t, err)

	folderID := folder.GetId()

	// Upload a small file into the folder to get a real file ID and content ID
	cwd, err := os.Getwd()
	require.NoError(t, err)

	repoRoot := filepath.Dir(cwd)
	imgPath := filepath.Join(repoRoot, "images", "testMedia", "IMG_3973.JPG")

	imgStat, err := os.Stat(imgPath)
	require.NoError(t, err)

	uploadInfo, _, err := adminClient.FilesAPI.StartUpload(t.Context()).Request(openapi.NewUploadParams{
		RootFolderID: openapi.PtrString(folderID),
		ChunkSize:    openapi.PtrInt32(int32(imgStat.Size())),
	}).Execute()
	require.NoError(t, err)

	uploadID := uploadInfo.GetUploadID()

	filesInfo, _, err := adminClient.FilesAPI.AddFilesToUpload(t.Context(), uploadID).Request(openapi.NewFilesParams{
		NewFiles: []openapi.NewFileParams{
			{
				NewFileName:    openapi.PtrString("IMG_3973.JPG"),
				ParentFolderID: openapi.PtrString(folderID),
				FileSize:       openapi.PtrInt32(int32(imgStat.Size())),
				IsDir:          openapi.PtrBool(false),
			},
		},
	}).Execute()
	require.NoError(t, err)

	uploadedFileIDs := filesInfo.GetFileIDs()
	require.Len(t, uploadedFileIDs, 1)

	uploadedFileID := uploadedFileIDs[0]

	imgBytes, err := os.ReadFile(imgPath)
	require.NoError(t, err)

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("chunk", "IMG_3973.JPG")
	require.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(imgBytes))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	chunkURL := fmt.Sprintf("%s/api/v1/upload/%s/file/%s", coreSetup.address, uploadID, uploadedFileID)
	chunkReq, err := http.NewRequestWithContext(t.Context(), http.MethodPut, chunkURL, &body)
	require.NoError(t, err)

	chunkReq.Header.Set("Authorization", "Bearer "+coreSetup.token)
	chunkReq.Header.Set("Content-Type", writer.FormDataContentType())
	chunkReq.Header.Set("Content-Range", fmt.Sprintf("bytes=0-%d/%d", len(imgBytes)-1, len(imgBytes)))

	chunkResp, err := http.DefaultClient.Do(chunkReq)
	require.NoError(t, err)

	defer func() { _ = chunkResp.Body.Close() }()

	require.Equal(t, http.StatusOK, chunkResp.StatusCode)

	_, err = adminClient.FilesAPI.GetUploadResult(t.Context(), uploadID).Execute()
	require.NoError(t, err)

	// Get the uploaded file to read its content ID
	uploadedFile, err := coreSetup.ctx.FileService.GetFileByID(coreSetup.ctx, uploadedFileID)
	require.NoError(t, err)

	contentID := uploadedFile.GetContentID()
	require.NotEmpty(t, contentID, "uploaded file should have a content ID")

	// Directly insert a media document into MongoDB for this content ID.
	// This bypasses image processing (which requires the agno Rust library)
	// but tests the exact media query + share access control path.
	syntheticMedia := &media_model.Media{
		ContentID:  media_model.ContentID(contentID),
		Owner:      "admin",
		MimeType:   "image/jpeg",
		FileIDs:    []string{uploadedFileID},
		Width:      1920,
		Height:     1080,
		PageCount:  1,
		CreateDate: time.Unix(1000, 0),
		Enabled:    true,
	}

	err = media_model.SaveMedia(coreSetup.ctx, syntheticMedia)
	require.NoError(t, err)

	// Verify admin can see the media
	adminMediaBatch, resp, err := adminClient.MediaAPI.GetMedia(t.Context()).
		FolderIDs([]string{folderID}).
		Page(0).
		Limit(200).
		Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Greater(t, adminMediaBatch.GetMediaCount(), int32(0), "admin should see the synthetic media")

	// Share the folder with the viewer user
	createdShare, _, err := adminClient.ShareAPI.CreateFileShare(t.Context()).Request(openapi.FileShareParams{
		FileID: openapi.PtrString(folderID),
		Public: openapi.PtrBool(false),
		Users:  []string{"timelineviewer"},
	}).Execute()
	require.NoError(t, err)

	shareID := createdShare.GetShareID()
	require.NotEmpty(t, shareID)

	// As the share recipient, query media in the shared folder using the shareID
	viewerMediaBatch, resp, err := viewerClient.MediaAPI.GetMedia(t.Context()).
		ShareID(shareID).
		FolderIDs([]string{folderID}).
		Page(0).
		Limit(200).
		Execute()
	require.NoError(t, err, "share recipient should be able to query media in shared folder")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Greater(t, viewerMediaBatch.GetMediaCount(), int32(0),
		"share recipient should see the same media as the folder owner")
	assert.Equal(t, adminMediaBatch.GetMediaCount(), viewerMediaBatch.GetMediaCount(),
		"share recipient and owner should see the same number of media items")
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
