package e2e_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFolder(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

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
	require.NoError(t, err, "Failed to start test server")

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

	path, err := wlfs.ParsePortable(fileInfo.GetPortablePath())
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
	require.NoError(t, err, "Failed to start test server")

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

	path, err := wlfs.ParsePortable(fileInfo.GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "getfile-test", path.Filename())

	// Test getting non-existent file
	_, resp, err = client.FilesAPI.GetFile(t.Context(), "non-existent-id").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSearchByFilename(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

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
	require.NoError(t, err, "Failed to start test server")

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

	path, err := wlfs.ParsePortable(fileInfo.GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "renamed-folder", path.Filename())
}

func TestMoveFiles(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

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

	path, err := wlfs.ParsePortable(targetFolderInfo.GetChildren()[0].GetPortablePath())
	require.NoError(t, err)

	assert.Equal(t, "folder-to-move", path.Filename())
}

func TestDeleteFiles(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

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

func TestDownloadFile(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user to find their home folder ID
	userInfo, _, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)

	homeID := userInfo.GetHomeID()
	require.NotEmpty(t, homeID)

	// Upload a real JPEG image
	cwd, err := os.Getwd()
	require.NoError(t, err)

	repoRoot := filepath.Dir(cwd)
	imgPath := filepath.Join(repoRoot, "images", "testMedia", "IMG_3973.JPG")

	imgStat, err := os.Stat(imgPath)
	require.NoError(t, err)

	uploadInfo, _, err := client.FilesAPI.StartUpload(t.Context()).Request(openapi.NewUploadParams{
		RootFolderID: openapi.PtrString(homeID),
		ChunkSize:    openapi.PtrInt32(int32(imgStat.Size())),
	}).Execute()
	require.NoError(t, err)

	uploadID := uploadInfo.GetUploadID()
	require.NotEmpty(t, uploadID)

	filesInfo, _, err := client.FilesAPI.AddFilesToUpload(t.Context(), uploadID).Request(openapi.NewFilesParams{
		NewFiles: []openapi.NewFileParams{
			{
				NewFileName:    openapi.PtrString("IMG_3973.JPG"),
				ParentFolderID: openapi.PtrString(homeID),
				FileSize:       openapi.PtrInt32(int32(imgStat.Size())),
				IsDir:          openapi.PtrBool(false),
			},
		},
	}).Execute()
	require.NoError(t, err)

	fileIDs := filesInfo.GetFileIDs()
	require.Len(t, fileIDs, 1)

	uploadedFileID := fileIDs[0]

	imgBytes, err := os.ReadFile(imgPath)
	require.NoError(t, err)

	// Upload the chunk via raw HTTP (the generated client doesn't set Content-Range)
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

	_, err = client.FilesAPI.GetUploadResult(t.Context(), uploadID).Execute()
	require.NoError(t, err)

	// Table-driven subtests for the generated client
	tests := []struct {
		name           string
		fileID         string
		format         *string
		quality        *int32
		isTakeout      *bool
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "basic download",
			fileID:         uploadedFileID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent file",
			fileID:         "non-existent-id",
			expectError:    true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid format",
			fileID:         uploadedFileID,
			format:         openapi.PtrString("not/real"),
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "format image/webp",
			fileID:         uploadedFileID,
			format:         openapi.PtrString("image/webp"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "isTakeout true",
			fileID:         uploadedFileID,
			isTakeout:      openapi.PtrBool(true),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "isTakeout false",
			fileID:         uploadedFileID,
			isTakeout:      openapi.PtrBool(false),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := client.FilesAPI.DownloadFile(t.Context(), tt.fileID)
			if tt.format != nil {
				req = req.Format(*tt.format)
			}

			if tt.quality != nil {
				req = req.Quality(*tt.quality)
			}

			if tt.isTakeout != nil {
				req = req.IsTakeout(*tt.isTakeout)
			}

			_, resp, err := req.Execute()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}

	// Raw HTTP request for invalid quality (client can't send non-numeric quality)
	t.Run("invalid quality string", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/files/%s/download?format=image/jpeg&quality=abc", coreSetup.address, uploadedFileID)
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+coreSetup.token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
