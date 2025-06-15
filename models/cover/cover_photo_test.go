package cover_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/require"

	. "github.com/ethanrous/weblens/models/cover"
)

func TestCoverPhoto_CRUD(t *testing.T) {
	ctx := db.SetupTestDB(t, CoverPhotoCollectionKey)

	folderId := "test-folder"
	coverPhotoId := "test-photo"
	coverPhotoId2 := "test-photo-2"

	// Test SetCoverPhoto (insert)
	cp, err := SetCoverPhoto(ctx, folderId, coverPhotoId)
	require.NoError(t, err)
	require.Equal(t, folderId, cp.FolderId)
	require.Equal(t, coverPhotoId, cp.CoverPhotoId)

	// Test GetCoverByFolderId (found)
	cp2, err := GetCoverByFolderId(ctx, folderId)
	require.NoError(t, err)
	require.Equal(t, coverPhotoId, cp2.CoverPhotoId)

	// Test SetCoverPhoto (replace)
	cp3, err := SetCoverPhoto(ctx, folderId, coverPhotoId2)
	require.NoError(t, err)
	require.Equal(t, coverPhotoId2, cp3.CoverPhotoId)

	// Test GetCoverByFolderId (after replace)
	cp4, err := GetCoverByFolderId(ctx, folderId)
	require.NoError(t, err)
	require.Equal(t, coverPhotoId2, cp4.CoverPhotoId)

	// Test UpsertCoverByFolderId (upsert new)
	folderId2 := "test-folder-2"
	err = UpsertCoverByFolderId(ctx, folderId2, coverPhotoId)
	require.NoError(t, err)
	cp5, err := GetCoverByFolderId(ctx, folderId2)
	require.NoError(t, err)
	require.Equal(t, coverPhotoId, cp5.CoverPhotoId)

	// Test DeleteCoverByFolderId (existing)
	err = DeleteCoverByFolderId(ctx, folderId)
	require.NoError(t, err)
	_, err = GetCoverByFolderId(ctx, folderId)
	require.Error(t, err)

	// Test DeleteCoverByFolderId (non-existing)
	err = DeleteCoverByFolderId(ctx, "nonexistent-folder")
	require.Error(t, err)
}
