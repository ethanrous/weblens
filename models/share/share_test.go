package share_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	testShareName = "Test Share"
	testUsername  = "testuser"
)

func createTestUser(suffix string) *user_model.User {
	return &user_model.User{
		ID:          primitive.NewObjectID(),
		Username:    testUsername + "_" + suffix,
		DisplayName: "Test User " + suffix,
	}
}

func TestFileShare_Creation(t *testing.T) {
	ctx := db.SetupTestDB(t, share_model.ShareCollectionKey, share_model.IndexModels...)

	t.Run("CreateBasicShare", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("basic")
		accessors := []*user_model.User{createTestUser("basic_accessor")}

		share, err := share_model.NewFileShare(ctx, fileID, owner, accessors, false, false, false)
		assert.NoError(t, err)
		assert.NotNil(t, share)

		err = share_model.SaveFileShare(ctx, share)
		assert.NoError(t, err)

		// Verify share was saved
		savedShare, err := share_model.GetShareByFileID(ctx, fileID)
		assert.NoError(t, err)
		assert.Equal(t, fileID, savedShare.FileID)
		assert.Equal(t, owner.GetUsername(), savedShare.Owner)
		assert.Contains(t, savedShare.Accessors, accessors[0].GetUsername())
	})

	t.Run("CreatePublicShare", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("public")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, true, false, false)
		assert.NoError(t, err)

		err = share_model.SaveFileShare(ctx, share)
		assert.NoError(t, err)

		savedShare, err := share_model.GetShareByFileID(ctx, fileID)
		assert.NoError(t, err)
		assert.True(t, savedShare.Public)
	})

	t.Run("CreateWormholeShare", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("wormhole")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, true, false)
		assert.NoError(t, err)

		err = share_model.SaveFileShare(ctx, share)
		assert.NoError(t, err)

		savedShare, err := share_model.GetShareByFileID(ctx, fileID)
		assert.NoError(t, err)
		assert.True(t, savedShare.Wormhole)
	})

	t.Run("CreateDuplicateShare", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("duplicate")
		share1, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share1)
		require.NoError(t, err)

		share2, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share2)
		assert.Error(t, err)
		assert.True(t, db.IsAlreadyExists(err), "Expected AlreadyExistsError, got: %v", err)
	})
}

func TestFileShare_Retrieval(t *testing.T) {
	ctx := db.SetupTestDB(t, share_model.ShareCollectionKey)

	t.Run("GetShareByID", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("retrieve_by_id")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		retrieved, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.Equal(t, share.ShareID, retrieved.ShareID)
		assert.Equal(t, share.FileID, retrieved.FileID)
	})

	t.Run("GetShareByFileID", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("retrieve_by_file")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		retrieved, err := share_model.GetShareByFileID(ctx, fileID)
		assert.NoError(t, err)
		assert.Equal(t, share.ShareID, retrieved.ShareID)
		assert.Equal(t, fileID, retrieved.FileID)
	})

	t.Run("GetNonexistentShare", func(t *testing.T) {
		_, err := share_model.GetShareByID(ctx, primitive.NewObjectID())
		assert.Error(t, err)
		assert.True(t, db.IsNotFound(err), "Expected NotFoundError, got: %v", err)

		_, err = share_model.GetShareByFileID(ctx, primitive.NewObjectID().Hex())
		assert.Error(t, err)
	})

	t.Run("GetSharedWithUser", func(t *testing.T) {
		owner := createTestUser("shared_with_owner")
		accessor := createTestUser("shared_with_accessor")

		// Create multiple shares with the same accessor
		numShares := 3
		for range numShares {
			fileID := primitive.NewObjectID().Hex()
			share, err := share_model.NewFileShare(ctx, fileID, owner, []*user_model.User{accessor}, false, false, false)
			require.NoError(t, err)
			err = share_model.SaveFileShare(ctx, share)
			require.NoError(t, err)
		}

		// Create a share with different accessor
		otherFileID := primitive.NewObjectID().Hex()
		otherShare, err := share_model.NewFileShare(ctx, otherFileID, owner, []*user_model.User{createTestUser("other")}, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, otherShare)
		require.NoError(t, err)

		shares, err := share_model.GetSharedWithUser(ctx, accessor.Username)
		assert.NoError(t, err)
		assert.Len(t, shares, numShares)

		for _, share := range shares {
			assert.Contains(t, share.Accessors, accessor.Username)
		}
	})
}

func TestFileShare_Updates(t *testing.T) {
	ctx := db.SetupTestDB(t, share_model.ShareCollectionKey)

	t.Run("SetPublic", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("public_update")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		err = share.SetPublic(ctx, true)
		assert.NoError(t, err)

		updated, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.True(t, updated.Public)
	})

	t.Run("AddUsers", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("add_users")
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		newUsers := []string{
			createTestUser("add1").Username,
			createTestUser("add2").Username,
		}
		for _, user := range newUsers {
			err = share.AddUser(ctx, user, share_model.NewPermissions())
			assert.NoError(t, err)
		}

		updated, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.Subset(t, updated.Accessors, newUsers)
	})

	t.Run("RemoveUsers", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("remove_users")
		initialUsers := []string{
			createTestUser("remove1").Username,
			createTestUser("remove2").Username,
			createTestUser("remove3").Username,
		}
		share, err := share_model.NewFileShare(ctx, fileID, owner, nil, false, false, false)
		require.NoError(t, err)

		share.Accessors = initialUsers

		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		usersToRemove := []string{initialUsers[0], initialUsers[1]}
		err = share.RemoveUsers(ctx, usersToRemove)
		assert.NoError(t, err)

		updated, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.NotContains(t, updated.Accessors, initialUsers[0])
		assert.NotContains(t, updated.Accessors, initialUsers[1])
		assert.Contains(t, updated.Accessors, initialUsers[2])
	})

	t.Run("Permissions", func(t *testing.T) {
		fileID := primitive.NewObjectID().Hex()
		owner := createTestUser("perm_owner")
		user := createTestUser("perm_user")
		share, err := share_model.NewFileShare(ctx, fileID, owner, []*user_model.User{user}, false, false, false)
		require.NoError(t, err)
		err = share_model.SaveFileShare(ctx, share)
		require.NoError(t, err)

		// Default permission should be download
		perms := share.GetUserPermissions(user.Username)
		assert.True(t, perms.CanDownload)
		assert.True(t, share.HasPermission(user.Username, share_model.SharePermissionDownload))
		assert.False(t, share.HasPermission(user.Username, share_model.SharePermissionEdit))

		// Set new permissions
		newPerms := &share_model.Permissions{CanDownload: true, CanEdit: true, CanDelete: false}
		err = share.SetUserPermissions(ctx, user.Username, newPerms)
		assert.NoError(t, err)
		updated, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.True(t, updated.HasPermission(user.Username, share_model.SharePermissionEdit))
		assert.True(t, updated.HasPermission(user.Username, share_model.SharePermissionDownload))
		assert.False(t, updated.HasPermission(user.Username, share_model.SharePermissionDelete))

		// Remove user and check permissions are gone
		err = updated.RemoveUsers(ctx, []string{user.Username})
		assert.NoError(t, err)
		final, err := share_model.GetShareByID(ctx, share.ShareID)
		assert.NoError(t, err)
		assert.Nil(t, final.GetUserPermissions(user.Username))
		assert.False(t, final.HasPermission(user.Username, share_model.SharePermissionEdit))
	})
}
