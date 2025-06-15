package share_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	. "github.com/ethanrous/weblens/models/share"
)

const (
	testShareName = "Test Share"
	testUsername  = "testuser"
)

func createTestUser(suffix string) *user_model.User {
	return &user_model.User{
		Id:          primitive.NewObjectID(),
		Username:    testUsername + "_" + suffix,
		DisplayName: "Test User " + suffix,
	}
}

func TestFileShare_Creation(t *testing.T) {
	ctx := db.SetupTestDB(t, ShareCollectionKey, IndexModels...)

	t.Run("CreateBasicShare", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("basic")
		accessors := []*user_model.User{createTestUser("basic_accessor")}

		share, err := NewFileShare(ctx, fileId, owner, accessors, false, false)
		assert.NoError(t, err)
		assert.NotNil(t, share)

		err = SaveFileShare(ctx, share)
		assert.NoError(t, err)

		// Verify share was saved
		savedShare, err := GetShareByFileId(ctx, fileId)
		assert.NoError(t, err)
		assert.Equal(t, fileId, savedShare.FileId)
		assert.Equal(t, owner.GetUsername(), savedShare.Owner)
		assert.Contains(t, savedShare.Accessors, accessors[0].GetUsername())
	})

	t.Run("CreatePublicShare", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("public")
		share, err := NewFileShare(ctx, fileId, owner, nil, true, false)
		assert.NoError(t, err)

		err = SaveFileShare(ctx, share)
		assert.NoError(t, err)

		savedShare, err := GetShareByFileId(ctx, fileId)
		assert.NoError(t, err)
		assert.True(t, savedShare.Public)
	})

	t.Run("CreateWormholeShare", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("wormhole")
		share, err := NewFileShare(ctx, fileId, owner, nil, false, true)
		assert.NoError(t, err)

		err = SaveFileShare(ctx, share)
		assert.NoError(t, err)

		savedShare, err := GetShareByFileId(ctx, fileId)
		assert.NoError(t, err)
		assert.True(t, savedShare.Wormhole)
	})

	t.Run("CreateDuplicateShare", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("duplicate")
		share1, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share1)
		require.NoError(t, err)

		share2, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share2)
		assert.Error(t, err)
		assert.True(t, db.IsAlreadyExists(err), "Expected AlreadyExistsError, got: %v", err)
	})
}

func TestFileShare_Retrieval(t *testing.T) {
	ctx := db.SetupTestDB(t, ShareCollectionKey)

	t.Run("GetShareById", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("retrieve_by_id")
		share, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		retrieved, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.Equal(t, share.ShareId, retrieved.ShareId)
		assert.Equal(t, share.FileId, retrieved.FileId)
	})

	t.Run("GetShareByFileId", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("retrieve_by_file")
		share, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		retrieved, err := GetShareByFileId(ctx, fileId)
		assert.NoError(t, err)
		assert.Equal(t, share.ShareId, retrieved.ShareId)
		assert.Equal(t, fileId, retrieved.FileId)
	})

	t.Run("GetNonexistentShare", func(t *testing.T) {
		_, err := GetShareById(ctx, primitive.NewObjectID())
		assert.Error(t, err)
		assert.True(t, db.IsNotFound(err), "Expected NotFoundError, got: %v", err)

		_, err = GetShareByFileId(ctx, primitive.NewObjectID().Hex())
		assert.Error(t, err)
	})

	t.Run("GetSharedWithUser", func(t *testing.T) {
		owner := createTestUser("shared_with_owner")
		accessor := createTestUser("shared_with_accessor")

		// Create multiple shares with the same accessor
		numShares := 3
		for range numShares {
			fileId := primitive.NewObjectID().Hex()
			share, err := NewFileShare(ctx, fileId, owner, []*user_model.User{accessor}, false, false)
			require.NoError(t, err)
			err = SaveFileShare(ctx, share)
			require.NoError(t, err)
		}

		// Create a share with different accessor
		otherFileId := primitive.NewObjectID().Hex()
		otherShare, err := NewFileShare(ctx, otherFileId, owner, []*user_model.User{createTestUser("other")}, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, otherShare)
		require.NoError(t, err)

		shares, err := GetSharedWithUser(ctx, accessor.Username)
		assert.NoError(t, err)
		assert.Len(t, shares, numShares)

		for _, share := range shares {
			assert.Contains(t, share.Accessors, accessor.Username)
		}
	})
}

func TestFileShare_Updates(t *testing.T) {
	ctx := db.SetupTestDB(t, ShareCollectionKey)

	t.Run("SetPublic", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("public_update")
		share, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		err = share.SetPublic(ctx, true)
		assert.NoError(t, err)

		updated, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.True(t, updated.Public)
	})

	t.Run("AddUsers", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("add_users")
		share, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		newUsers := []string{
			createTestUser("add1").Username,
			createTestUser("add2").Username,
		}
		for _, user := range newUsers {
			err = share.AddUser(ctx, user, NewPermissions())
			assert.NoError(t, err)
		}

		updated, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.Subset(t, updated.Accessors, newUsers)
	})

	t.Run("RemoveUsers", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("remove_users")
		initialUsers := []string{
			createTestUser("remove1").Username,
			createTestUser("remove2").Username,
			createTestUser("remove3").Username,
		}
		share, err := NewFileShare(ctx, fileId, owner, nil, false, false)
		require.NoError(t, err)
		share.Accessors = initialUsers
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		usersToRemove := []string{initialUsers[0], initialUsers[1]}
		err = share.RemoveUsers(ctx, usersToRemove)
		assert.NoError(t, err)

		updated, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.NotContains(t, updated.Accessors, initialUsers[0])
		assert.NotContains(t, updated.Accessors, initialUsers[1])
		assert.Contains(t, updated.Accessors, initialUsers[2])
	})

	t.Run("Permissions", func(t *testing.T) {
		fileId := primitive.NewObjectID().Hex()
		owner := createTestUser("perm_owner")
		user := createTestUser("perm_user")
		share, err := NewFileShare(ctx, fileId, owner, []*user_model.User{user}, false, false)
		require.NoError(t, err)
		err = SaveFileShare(ctx, share)
		require.NoError(t, err)

		// Default permission should be download
		perms := share.GetUserPermissions(user.Username)
		assert.True(t, perms.CanDownload)
		assert.True(t, share.HasPermission(user.Username, SharePermissionDownload))
		assert.False(t, share.HasPermission(user.Username, SharePermissionEdit))

		// Set new permissions
		newPerms := &Permissions{CanDownload: true, CanEdit: true, CanDelete: false}
		err = share.SetUserPermissions(ctx, user.Username, newPerms)
		assert.NoError(t, err)
		updated, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.True(t, updated.HasPermission(user.Username, SharePermissionEdit))
		assert.True(t, updated.HasPermission(user.Username, SharePermissionDownload))
		assert.False(t, updated.HasPermission(user.Username, SharePermissionDelete))

		// Remove user and check permissions are gone
		err = updated.RemoveUsers(ctx, []string{user.Username})
		assert.NoError(t, err)
		final, err := GetShareById(ctx, share.ShareId)
		assert.NoError(t, err)
		assert.Nil(t, final.GetUserPermissions(user.Username))
		assert.False(t, final.HasPermission(user.Username, SharePermissionEdit))
	})
}
