package history_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetLifetimes_RestoredFileIncluded(t *testing.T) {
	ctx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	now := time.Now()
	fileID := primitive.NewObjectID().Hex()
	eventID := primitive.NewObjectID().Hex()
	fp := wlfs.BuildFilePath("USERS", "testuser/restored-file.txt")

	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	require.NoError(t, err)

	// Insert a fileRestore action
	_, err = col.InsertOne(ctx, bson.M{
		"actionType": "fileRestore",
		"fileID":     fileID,
		"filepath":   fp.ToPortable(),
		"eventID":    eventID,
		"towerID":    "test-tower",
		"timestamp":  now,
	})
	require.NoError(t, err)

	lifetimes, err := history.GetLifetimes(ctx, history.GetLifetimesOptions{
		ActiveOnly: true,
		TowerID:    "test-tower",
	})
	require.NoError(t, err)

	require.Equal(t, 1, len(lifetimes), "restored file should appear in active lifetimes")
	assert.Equal(t, fileID, lifetimes[0].ID)
}

func TestGetLifetimes_MultipleRestoredFilesNotDeduped(t *testing.T) {
	ctx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	now := time.Now()
	eventID := primitive.NewObjectID().Hex()

	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	require.NoError(t, err)

	// Insert 3 restored files from the same event, each with unique fileID and filepath
	files := []struct {
		fileID string
		path   string
	}{
		{primitive.NewObjectID().Hex(), "testuser/file-a.txt"},
		{primitive.NewObjectID().Hex(), "testuser/file-b.txt"},
		{primitive.NewObjectID().Hex(), "testuser/file-c.txt"},
	}

	for _, f := range files {
		fp := wlfs.BuildFilePath("USERS", f.path)
		_, err = col.InsertOne(ctx, bson.M{
			"actionType": "fileRestore",
			"fileID":     f.fileID,
			"filepath":   fp.ToPortable(),
			"eventID":    eventID,
			"towerID":    "test-tower",
			"timestamp":  now,
		})
		require.NoError(t, err)
	}

	lifetimes, err := history.GetLifetimes(ctx, history.GetLifetimesOptions{
		ActiveOnly: true,
		TowerID:    "test-tower",
	})
	require.NoError(t, err)

	require.Equal(t, 3, len(lifetimes), "all 3 restored files should appear, not deduped to 1")

	gotIDs := make(map[string]bool)
	for _, lt := range lifetimes {
		gotIDs[lt.ID] = true
	}

	for _, f := range files {
		assert.True(t, gotIDs[f.fileID], "missing restored file %s", f.fileID)
	}
}

func TestGetLifetimes_MovedFileNotDroppedByRestoreAtSamePath(t *testing.T) {
	ctx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	require.NoError(t, err)

	t0 := time.Now()
	t1 := t0.Add(time.Second)
	t2 := t0.Add(2 * time.Second)

	originalFileID := primitive.NewObjectID().Hex()
	restoredFileID := primitive.NewObjectID().Hex()
	createPath := wlfs.BuildFilePath("USERS", "admin/photo.jpg")
	movedPath := wlfs.BuildFilePath("USERS", "admin/.user_trash/photo.jpg")

	// File created at photo.jpg
	_, err = col.InsertOne(ctx, bson.M{
		"actionType": "fileCreate", "fileID": originalFileID,
		"filepath": createPath.ToPortable(), "towerID": "test-tower",
		"eventID": primitive.NewObjectID().Hex(), "timestamp": t0,
	})
	require.NoError(t, err)

	// File moved to trash
	_, err = col.InsertOne(ctx, bson.M{
		"actionType": "fileMove", "fileID": originalFileID,
		"originPath": createPath.ToPortable(), "destinationPath": movedPath.ToPortable(),
		"towerID": "test-tower", "eventID": primitive.NewObjectID().Hex(), "timestamp": t1,
	})
	require.NoError(t, err)

	// New file restored at the same original path
	_, err = col.InsertOne(ctx, bson.M{
		"actionType": "fileRestore", "fileID": restoredFileID,
		"filepath": createPath.ToPortable(), "towerID": "test-tower",
		"eventID": primitive.NewObjectID().Hex(), "timestamp": t2,
	})
	require.NoError(t, err)

	lifetimes, err := history.GetLifetimes(ctx, history.GetLifetimesOptions{
		ActiveOnly: true,
		TowerID:    "test-tower",
	})
	require.NoError(t, err)

	require.Equal(t, 2, len(lifetimes), "both the moved file and the restored file should appear")

	gotIDs := make(map[string]bool)
	for _, lt := range lifetimes {
		gotIDs[lt.ID] = true
	}

	assert.True(t, gotIDs[originalFileID], "moved-to-trash file should still be in lifetimes")
	assert.True(t, gotIDs[restoredFileID], "restored file should be in lifetimes")
}

func TestDoesFileExistInHistory_Restore(t *testing.T) {
	ctx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	now := time.Now()
	fileID := primitive.NewObjectID().Hex()
	fp := wlfs.BuildFilePath("USERS", "testuser/restored.txt")

	col, err := db.GetCollection[any](ctx, history.FileHistoryCollectionKey)
	require.NoError(t, err)

	_, err = col.InsertOne(ctx, bson.M{
		"actionType": "fileRestore",
		"fileID":     fileID,
		"filepath":   fp.ToPortable(),
		"eventID":    primitive.NewObjectID().Hex(),
		"towerID":    "test-tower",
		"timestamp":  now,
	})
	require.NoError(t, err)

	action, err := history.DoesFileExistInHistory(ctx, fp)
	require.NoError(t, err)
	require.NotNil(t, action, "DoesFileExistInHistory should find fileRestore action")
	assert.Equal(t, fileID, action.FileID)
}

func TestDoesFileExistInHistory_NoMatch(t *testing.T) {
	ctx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	fp := wlfs.BuildFilePath("USERS", "testuser/nonexistent.txt")

	action, err := history.DoesFileExistInHistory(ctx, fp)
	require.NoError(t, err)
	assert.Nil(t, action, "should return nil when no history exists")
}
