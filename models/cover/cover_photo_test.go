package cover_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/cover"
	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/require"
)

func TestCoverPhoto_CRUD(t *testing.T) {
	ctx := db.SetupTestDB(t, cover.CoverPhotoCollectionKey)

	folderID := "test-folder"
	coverPhotoID := "test-photo"
	coverPhotoID2 := "test-photo-2"

	// Test SetCoverPhoto (insert)
	cp, err := cover.SetCoverPhoto(ctx, folderID, coverPhotoID)
	require.NoError(t, err)
	require.Equal(t, folderID, cp.FolderID)
	require.Equal(t, coverPhotoID, cp.CoverPhotoID)

	// Test GetCoverByFolderID (found)
	cp2, err := cover.GetCoverByFolderID(ctx, folderID)
	require.NoError(t, err)
	require.Equal(t, coverPhotoID, cp2.CoverPhotoID)

	// Test SetCoverPhoto (replace)
	cp3, err := cover.SetCoverPhoto(ctx, folderID, coverPhotoID2)
	require.NoError(t, err)
	require.Equal(t, coverPhotoID2, cp3.CoverPhotoID)

	// Test GetCoverByFolderID (after replace)
	cp4, err := cover.GetCoverByFolderID(ctx, folderID)
	require.NoError(t, err)
	require.Equal(t, coverPhotoID2, cp4.CoverPhotoID)

	// Test UpsertCoverByFolderID (upsert new)
	folderID2 := "test-folder-2"
	err = cover.UpsertCoverByFolderID(ctx, folderID2, coverPhotoID)
	require.NoError(t, err)
	cp5, err := cover.GetCoverByFolderID(ctx, folderID2)
	require.NoError(t, err)
	require.Equal(t, coverPhotoID, cp5.CoverPhotoID)

	// Test DeleteCoverByFolderID (existing)
	err = cover.DeleteCoverByFolderID(ctx, folderID)
	require.NoError(t, err)
	_, err = cover.GetCoverByFolderID(ctx, folderID)
	require.Error(t, err)

	// Test DeleteCoverByFolderID (non-existing)
	err = cover.DeleteCoverByFolderID(ctx, "nonexistent-folder")
	require.Error(t, err)
}

func TestGetCoversByFolderIDs(t *testing.T) {
	ctx := db.SetupTestDB(t, cover.CoverPhotoCollectionKey)

	// Set up covers for multiple folders
	_, err := cover.SetCoverPhoto(ctx, "folder-a", "photo-a")
	require.NoError(t, err)
	_, err = cover.SetCoverPhoto(ctx, "folder-b", "photo-b")
	require.NoError(t, err)

	// Batch fetch including a folder with no cover
	covers, err := cover.GetCoversByFolderIDs(ctx, "folder-a", "folder-b", "folder-c")
	require.NoError(t, err)
	require.Len(t, covers, 2)

	coverMap := make(map[string]string)
	for _, c := range covers {
		coverMap[c.FolderID] = c.CoverPhotoID
	}

	require.Equal(t, "photo-a", coverMap["folder-a"])
	require.Equal(t, "photo-b", coverMap["folder-b"])

	// Empty input returns empty result
	covers, err = cover.GetCoversByFolderIDs(ctx)
	require.NoError(t, err)
	require.Empty(t, covers)
}
