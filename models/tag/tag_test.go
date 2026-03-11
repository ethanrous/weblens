package tag_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTag_CRUD(t *testing.T) {
	ctx := db.SetupTestDB(t, tag.TagCollectionKey, tag.IndexModels...)

	t.Run("CreateTag", func(t *testing.T) {
		created, err := tag.CreateTag(ctx, "Vacation", "#e74c3c", "alice")
		require.NoError(t, err)
		assert.Equal(t, "Vacation", created.Name)
		assert.Equal(t, "#e74c3c", created.Color)
		assert.Equal(t, "alice", created.Owner)
		assert.NotZero(t, created.TagID)
		assert.Empty(t, created.FileIDs)
		assert.False(t, created.Created.IsZero())
		assert.False(t, created.Updated.IsZero())
	})

	t.Run("GetTagByID", func(t *testing.T) {
		created, err := tag.CreateTag(ctx, "Work", "#3498db", "alice")
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Equal(t, created.Name, fetched.Name)
		assert.Equal(t, created.Color, fetched.Color)
		assert.Equal(t, created.Owner, fetched.Owner)
	})

	t.Run("GetTagByID_NotFound", func(t *testing.T) {
		_, err := tag.GetTagByID(ctx, primitive.NewObjectID())
		assert.Error(t, err)
	})

	t.Run("GetTagsByOwner", func(t *testing.T) {
		_, err := tag.CreateTag(ctx, "Personal", "#2ecc71", "bob")
		require.NoError(t, err)
		_, err = tag.CreateTag(ctx, "Urgent", "#e67e22", "bob")
		require.NoError(t, err)

		tags, err := tag.GetTagsByOwner(ctx, "bob")
		require.NoError(t, err)
		assert.Len(t, tags, 2)

		names := []string{tags[0].Name, tags[1].Name}
		assert.Contains(t, names, "Personal")
		assert.Contains(t, names, "Urgent")
	})

	t.Run("GetTagsByOwner_Empty", func(t *testing.T) {
		tags, err := tag.GetTagsByOwner(ctx, "nonexistent-user")
		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("UpdateTag_Name", func(t *testing.T) {
		created, err := tag.CreateTag(ctx, "OldName", "#000000", "alice")
		require.NoError(t, err)

		err = tag.UpdateTag(ctx, created.TagID, "NewName", "")
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Equal(t, "NewName", fetched.Name)
		assert.Equal(t, "#000000", fetched.Color) // color unchanged
	})

	t.Run("UpdateTag_Color", func(t *testing.T) {
		created, err := tag.CreateTag(ctx, "ColorTest", "#111111", "alice")
		require.NoError(t, err)

		err = tag.UpdateTag(ctx, created.TagID, "", "#ffffff")
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Equal(t, "ColorTest", fetched.Name) // name unchanged
		assert.Equal(t, "#ffffff", fetched.Color)
	})

	t.Run("UpdateTag_NotFound", func(t *testing.T) {
		err := tag.UpdateTag(ctx, primitive.NewObjectID(), "Name", "#000")
		assert.Error(t, err)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})

	t.Run("DeleteTag", func(t *testing.T) {
		created, err := tag.CreateTag(ctx, "ToDelete", "#333333", "carol")
		require.NoError(t, err)

		err = tag.DeleteTag(ctx, created.TagID)
		require.NoError(t, err)

		_, err = tag.GetTagByID(ctx, created.TagID)
		assert.Error(t, err)
	})

	t.Run("DeleteTag_NotFound", func(t *testing.T) {
		err := tag.DeleteTag(ctx, primitive.NewObjectID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})
}

func TestTag_DuplicateName(t *testing.T) {
	ctx := db.SetupTestDB(t, tag.TagCollectionKey, tag.IndexModels...)

	_, err := tag.CreateTag(ctx, "Photos", "#e74c3c", "alice")
	require.NoError(t, err)

	// Same name, same owner -> should fail (unique index on owner+name)
	_, err = tag.CreateTag(ctx, "Photos", "#3498db", "alice")
	assert.Error(t, err)
	assert.True(t, db.IsAlreadyExists(err), "Expected AlreadyExistsError, got: %v", err)

	// Same name, different owner -> should succeed
	_, err = tag.CreateTag(ctx, "Photos", "#3498db", "bob")
	assert.NoError(t, err)
}

func TestTag_FileOperations(t *testing.T) {
	ctx := db.SetupTestDB(t, tag.TagCollectionKey, tag.IndexModels...)

	created, err := tag.CreateTag(ctx, "Tagged", "#e74c3c", "alice")
	require.NoError(t, err)

	fileID1 := "file-aaa"
	fileID2 := "file-bbb"
	fileID3 := "file-ccc"

	t.Run("AddFilesToTag", func(t *testing.T) {
		err := tag.AddFilesToTag(ctx, created.TagID, []string{fileID1, fileID2})
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Len(t, fetched.FileIDs, 2)
		assert.Contains(t, fetched.FileIDs, fileID1)
		assert.Contains(t, fetched.FileIDs, fileID2)
	})

	t.Run("AddFilesToTag_Deduplicate", func(t *testing.T) {
		// Adding same file again should not duplicate
		err := tag.AddFilesToTag(ctx, created.TagID, []string{fileID1})
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Len(t, fetched.FileIDs, 2) // still 2, not 3
	})

	t.Run("AddFilesToTag_NotFound", func(t *testing.T) {
		err := tag.AddFilesToTag(ctx, primitive.NewObjectID(), []string{fileID1})
		assert.Error(t, err)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})

	t.Run("RemoveFilesFromTag", func(t *testing.T) {
		err := tag.RemoveFilesFromTag(ctx, created.TagID, []string{fileID1})
		require.NoError(t, err)

		fetched, err := tag.GetTagByID(ctx, created.TagID)
		require.NoError(t, err)
		assert.Len(t, fetched.FileIDs, 1)
		assert.Contains(t, fetched.FileIDs, fileID2)
		assert.NotContains(t, fetched.FileIDs, fileID1)
	})

	t.Run("RemoveFilesFromTag_NotFound", func(t *testing.T) {
		err := tag.RemoveFilesFromTag(ctx, primitive.NewObjectID(), []string{fileID1})
		assert.Error(t, err)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})

	t.Run("GetTagsForFile", func(t *testing.T) {
		// Create a second tag and add file3 to both
		tag2, err := tag.CreateTag(ctx, "SecondTag", "#2ecc71", "alice")
		require.NoError(t, err)

		err = tag.AddFilesToTag(ctx, created.TagID, []string{fileID3})
		require.NoError(t, err)
		err = tag.AddFilesToTag(ctx, tag2.TagID, []string{fileID3})
		require.NoError(t, err)

		tags, err := tag.GetTagsForFile(ctx, fileID3)
		require.NoError(t, err)
		assert.Len(t, tags, 2)
	})

	t.Run("GetTagsForFile_NoTags", func(t *testing.T) {
		tags, err := tag.GetTagsForFile(ctx, "untagged-file")
		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("RemoveFileFromAllTags", func(t *testing.T) {
		// fileID3 is in both tags from earlier
		err := tag.RemoveFileFromAllTags(ctx, fileID3)
		require.NoError(t, err)

		tags, err := tag.GetTagsForFile(ctx, fileID3)
		require.NoError(t, err)
		assert.Empty(t, tags)
	})
}
