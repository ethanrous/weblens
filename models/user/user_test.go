package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	testUsername    = "testuser"
	testPassword    = "TestP@ssw0rd123"
	testDisplayName = "Test User"
)

func getTestCtx(t *testing.T) context.Context {
	ctx := db.SetupTestDB(t, UserCollectionKey)
	ctx = context.WithValue(ctx, crypto.BcryptDifficultyCtxKey, 4) // Set bcrypt difficulty for testing

	return ctx
}

func TestUser_Creation(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("CreateValidUser", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		assert.NoError(t, err)

		// Verify password was hashed
		assert.NotEqual(t, testPassword, user.Password)
		assert.True(t, crypto.VerifyUserPassword(testPassword, user.Password) == nil)
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		user1 := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_dupe",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		user2 := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_dupe",
			Password:    "different_password",
			DisplayName: "Different Name",
		}

		err := SaveUser(ctx, user1)
		assert.NoError(t, err)

		err = SaveUser(ctx, user2)
		assert.Error(t, err)
	})

	t.Run("CreateUserWithInvalidUsername", func(t *testing.T) {
		invalidUsernames := []string{
			"",                                // Empty
			"a",                               // Too short
			"user@example.com",                // Contains @
			"user with spaces",                // Contains spaces
			"user#special",                    // Special characters
			"verylongusername123456789012345", // Too long
		}

		for _, username := range invalidUsernames {
			user := &User{
				Id:          primitive.NewObjectID(),
				Username:    username,
				Password:    testPassword,
				DisplayName: testDisplayName,
			}

			err := SaveUser(ctx, user)
			assert.Error(t, err, "Should fail for username: %s", username)
		}
	})

	t.Run("CreateUserWithInvalidPassword", func(t *testing.T) {
		invalidPasswords := []string{
			"",        // Empty
			"t1ny",    // Too short
			"nodigit", // No numbers
		}

		for i, password := range invalidPasswords {
			user := &User{
				Id:          primitive.NewObjectID(),
				Username:    testUsername + fmt.Sprintf("_invalid_pass_%d", i),
				Password:    password,
				DisplayName: testDisplayName,
			}

			err := SaveUser(ctx, user)
			assert.Error(t, err, "Should fail for password: %s", password)
		}
	})
}

func TestUser_Authentication(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("ValidLogin", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.True(t, retrievedUser.CheckLogin(testPassword))
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_invalid",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.False(t, retrievedUser.CheckLogin("wrongpassword"))
	})

	t.Run("InactiveUser", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_inactive",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   false,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.False(t, retrievedUser.CheckLogin(testPassword))
	})
}

func TestUser_Permissions(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("PermissionLevels", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			UserPerms:   UserPermissionBasic,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		assert.False(t, user.IsAdmin())
		assert.False(t, user.IsOwner())
		assert.False(t, user.IsSystemUser())

		user.UserPerms = UserPermissionAdmin
		assert.True(t, user.IsAdmin())
		assert.False(t, user.IsOwner())
		assert.False(t, user.IsSystemUser())

		user.UserPerms = UserPermissionOwner
		assert.True(t, user.IsAdmin())
		assert.True(t, user.IsOwner())
		assert.False(t, user.IsSystemUser())

		user.UserPerms = UserPermissionSystem
		assert.True(t, user.IsAdmin())
		assert.True(t, user.IsOwner())
		assert.True(t, user.IsSystemUser())
	})

	t.Run("PermissionUpdates", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_update",
			Password:    testPassword,
			DisplayName: testDisplayName,
			UserPerms:   UserPermissionBasic,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		err = user.UpdatePermissionLevel(ctx, UserPermissionAdmin)
		assert.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, UserPermissionAdmin, retrievedUser.UserPerms)
	})
}

func TestUser_Retrieval(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("GetUserByUsername", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.DisplayName, retrieved.DisplayName)
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Drop collection to ensure a clean state
		col, err := db.GetCollection[any](ctx, UserCollectionKey)
		require.NoError(t, err)
		err = col.Drop(ctx)
		require.NoError(t, err)

		// Create multiple users
		numUsers := 5
		for i := range numUsers {
			user := &User{
				Id:          primitive.NewObjectID(),
				Username:    fmt.Sprintf("%s_%d", testUsername, i),
				Password:    testPassword,
				DisplayName: fmt.Sprintf("%s %d", testDisplayName, i),
				Activated:   true,
			}
			err := SaveUser(ctx, user)
			require.NoError(t, err)
		}

		users, err := GetAllUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, numUsers)
	})

	t.Run("GetServerOwner", func(t *testing.T) {
		owner := &User{
			Id:          primitive.NewObjectID(),
			Username:    "owner",
			Password:    testPassword,
			DisplayName: "Server Owner",
			UserPerms:   UserPermissionOwner,
			Activated:   true,
		}

		err := SaveUser(ctx, owner)
		require.NoError(t, err)

		retrievedOwner, err := GetServerOwner(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedOwner)
		assert.Equal(t, owner.Username, retrievedOwner.Username)
		assert.Equal(t, UserPermissionOwner, retrievedOwner.UserPerms)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		_, err := GetUserByUsername(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

func TestUser_Updates(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("UpdatePassword", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		newPassword := "NewP@ssw0rd123"
		err = user.UpdatePassword(ctx, newPassword)
		assert.NoError(t, err)

		// Verify password was updated and hashed
		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.True(t, crypto.VerifyUserPassword(newPassword, retrievedUser.Password) == nil)
		assert.False(t, crypto.VerifyUserPassword(testPassword, retrievedUser.Password) == nil)
	})

	t.Run("UpdateDisplayName", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_display",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		newDisplayName := "Updated Name"
		err = user.UpdateDisplayName(ctx, newDisplayName)
		assert.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, newDisplayName, retrievedUser.DisplayName)
	})

	t.Run("UpdateActivationStatus", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_activation",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   false,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		err = user.UpdateActivationStatus(ctx, true)
		assert.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.True(t, retrievedUser.Activated)
	})
}

func TestUser_Delete(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("DeleteUser", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		err = user.Delete(ctx)
		assert.NoError(t, err)

		// Verify user was deleted
		_, err = GetUserByUsername(ctx, user.Username)
		assert.Error(t, err)
	})

	t.Run("DeleteAllUsers", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 5; i++ {
			user := &User{
				Id:          primitive.NewObjectID(),
				Username:    fmt.Sprintf("%s_%d", testUsername, i),
				Password:    testPassword,
				DisplayName: fmt.Sprintf("%s %d", testDisplayName, i),
			}
			err := SaveUser(ctx, user)
			require.NoError(t, err)
		}

		err := DeleteAllUsers(ctx)
		assert.NoError(t, err)

		users, err := GetAllUsers(ctx)
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUser_Search(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("SearchByUsername", func(t *testing.T) {
		// Create users with similar usernames
		usernames := []string{"john_doe", "john_smith", "jane_doe", "bob_smith"}
		for _, username := range usernames {
			user := &User{
				Id:          primitive.NewObjectID(),
				Username:    username,
				Password:    testPassword,
				DisplayName: "Test User",
			}
			err := SaveUser(ctx, user)
			require.NoError(t, err)
		}

		// Test partial matches
		results, err := SearchByUsername(ctx, "john")
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		results, err = SearchByUsername(ctx, "doe")
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		// Test exact match
		results, err = SearchByUsername(ctx, "john_doe")
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// Test no matches
		results, err = SearchByUsername(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestUser_HomeAndTrash(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("UpdateHomeId", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		homeId := primitive.NewObjectID().Hex()
		err = user.UpdateHomeId(ctx, homeId)
		assert.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, homeId, retrievedUser.HomeId)
	})

	t.Run("UpdateTrashId", func(t *testing.T) {
		user := &User{
			Id:          primitive.NewObjectID(),
			Username:    testUsername + "_trash",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := SaveUser(ctx, user)
		require.NoError(t, err)

		trashId := primitive.NewObjectID().Hex()
		err = user.UpdateTrashId(ctx, trashId)
		assert.NoError(t, err)

		retrievedUser, err := GetUserByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, retrievedUser.Username)
		assert.Equal(t, trashId, retrievedUser.TrashId)
	})
}
