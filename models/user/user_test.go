package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/cryptography"
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
	ctx := db.SetupTestDB(t, user.UserCollectionKey)
	ctx = context.WithValue(ctx, cryptography.BcryptDifficultyCtxKey, 4) // Set bcrypt difficulty for testing

	return ctx
}

func TestUser_Creation(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("CreateValidUser", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		assert.NoError(t, err)

		// Verify password was hashed
		assert.NotEqual(t, testPassword, usr.Password)
		assert.True(t, cryptography.VerifyUserPassword(testPassword, usr.Password) == nil)
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		user1 := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_dupe",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		user2 := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_dupe",
			Password:    "different_password",
			DisplayName: "Different Name",
		}

		err := user.SaveUser(ctx, user1)
		assert.NoError(t, err)

		err = user.SaveUser(ctx, user2)
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
			usr := &user.User{
				ID:          primitive.NewObjectID(),
				Username:    username,
				Password:    testPassword,
				DisplayName: testDisplayName,
			}

			err := user.SaveUser(ctx, usr)
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
			usr := &user.User{
				ID:          primitive.NewObjectID(),
				Username:    testUsername + fmt.Sprintf("_invalid_pass_%d", i),
				Password:    password,
				DisplayName: testDisplayName,
			}

			err := user.SaveUser(ctx, usr)
			assert.Error(t, err, "Should fail for password: %s", password)
		}
	})
}

func TestUser_Authentication(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("ValidLogin", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.True(t, retrievedUser.CheckLogin(testPassword))
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_invalid",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.False(t, retrievedUser.CheckLogin("wrongpassword"))
	})

	t.Run("InactiveUser", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_inactive",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   false,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.False(t, retrievedUser.CheckLogin(testPassword))
	})
}

func TestUser_Permissions(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("PermissionLevels", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			UserPerms:   user.UserPermissionBasic,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		assert.False(t, usr.IsAdmin())
		assert.False(t, usr.IsOwner())
		assert.False(t, usr.IsSystemUser())

		usr.UserPerms = user.UserPermissionAdmin
		assert.True(t, usr.IsAdmin())
		assert.False(t, usr.IsOwner())
		assert.False(t, usr.IsSystemUser())

		usr.UserPerms = user.UserPermissionOwner
		assert.True(t, usr.IsAdmin())
		assert.True(t, usr.IsOwner())
		assert.False(t, usr.IsSystemUser())

		usr.UserPerms = user.UserPermissionSystem
		assert.True(t, usr.IsAdmin())
		assert.True(t, usr.IsOwner())
		assert.True(t, usr.IsSystemUser())
	})

	t.Run("PermissionUpdates", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_update",
			Password:    testPassword,
			DisplayName: testDisplayName,
			UserPerms:   user.UserPermissionBasic,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		err = usr.UpdatePermissionLevel(ctx, user.UserPermissionAdmin)
		assert.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.UserPermissionAdmin, retrievedUser.UserPerms)
	})
}

func TestUser_Retrieval(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("GetUserByUsername", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		retrieved, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.Equal(t, usr.Username, retrieved.Username)
		assert.Equal(t, usr.DisplayName, retrieved.DisplayName)
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Drop collection to ensure a clean state
		col, err := db.GetCollection[any](ctx, user.UserCollectionKey)
		require.NoError(t, err)
		err = col.Drop(ctx)
		require.NoError(t, err)

		// Create multiple users
		numUsers := 5
		for i := range numUsers {
			usr := &user.User{
				ID:          primitive.NewObjectID(),
				Username:    fmt.Sprintf("%s_%d", testUsername, i),
				Password:    testPassword,
				DisplayName: fmt.Sprintf("%s %d", testDisplayName, i),
				Activated:   true,
			}
			err := user.SaveUser(ctx, usr)
			require.NoError(t, err)
		}

		users, err := user.GetAllUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, numUsers)
	})

	t.Run("GetServerOwner", func(t *testing.T) {
		owner := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    "owner",
			Password:    testPassword,
			DisplayName: "Server Owner",
			UserPerms:   user.UserPermissionOwner,
			Activated:   true,
		}

		err := user.SaveUser(ctx, owner)
		require.NoError(t, err)

		retrievedOwner, err := user.GetServerOwner(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedOwner)
		assert.Equal(t, owner.Username, retrievedOwner.Username)
		assert.Equal(t, user.UserPermissionOwner, retrievedOwner.UserPerms)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		_, err := user.GetUserByUsername(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

func TestUser_Updates(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("UpdatePassword", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   true,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		newPassword := "NewP@ssw0rd123"
		err = usr.UpdatePassword(ctx, newPassword)
		assert.NoError(t, err)

		// Verify password was updated and hashed
		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.True(t, cryptography.VerifyUserPassword(newPassword, retrievedUser.Password) == nil)
		assert.False(t, cryptography.VerifyUserPassword(testPassword, retrievedUser.Password) == nil)
	})

	t.Run("UpdateDisplayName", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_display",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		newDisplayName := "Updated Name"
		err = usr.UpdateDisplayName(ctx, newDisplayName)
		assert.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.Equal(t, newDisplayName, retrievedUser.DisplayName)
	})

	t.Run("UpdateActivationStatus", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_activation",
			Password:    testPassword,
			DisplayName: testDisplayName,
			Activated:   false,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		err = usr.UpdateActivationStatus(ctx, true)
		assert.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.True(t, retrievedUser.Activated)
	})
}

func TestUser_Delete(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("DeleteUser", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		err = usr.Delete(ctx)
		assert.NoError(t, err)

		// Verify user was deleted
		_, err = user.GetUserByUsername(ctx, usr.Username)
		assert.Error(t, err)
	})

	t.Run("DeleteAllUsers", func(t *testing.T) {
		// Create multiple users
		for i := range 5 {
			usr := &user.User{
				ID:          primitive.NewObjectID(),
				Username:    fmt.Sprintf("%s_%d", testUsername, i),
				Password:    testPassword,
				DisplayName: fmt.Sprintf("%s %d", testDisplayName, i),
			}
			err := user.SaveUser(ctx, usr)
			require.NoError(t, err)
		}

		err := user.DeleteAllUsers(ctx)
		assert.NoError(t, err)

		users, err := user.GetAllUsers(ctx)
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
			usr := &user.User{
				ID:          primitive.NewObjectID(),
				Username:    username,
				Password:    testPassword,
				DisplayName: "Test User",
			}
			err := user.SaveUser(ctx, usr)
			require.NoError(t, err)
		}

		// Test partial matches
		results, err := user.SearchByUsername(ctx, "john")
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		results, err = user.SearchByUsername(ctx, "doe")
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		// Test exact match
		results, err = user.SearchByUsername(ctx, "john_doe")
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// Test no matches
		results, err = user.SearchByUsername(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestUser_HomeAndTrash(t *testing.T) {
	ctx := getTestCtx(t)

	t.Run("UpdateHomeID", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername,
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		homeID := primitive.NewObjectID().Hex()
		err = usr.UpdateHomeID(ctx, homeID)
		assert.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.Equal(t, homeID, retrievedUser.HomeID)
	})

	t.Run("UpdateTrashID", func(t *testing.T) {
		usr := &user.User{
			ID:          primitive.NewObjectID(),
			Username:    testUsername + "_trash",
			Password:    testPassword,
			DisplayName: testDisplayName,
		}

		err := user.SaveUser(ctx, usr)
		require.NoError(t, err)

		trashID := primitive.NewObjectID().Hex()
		err = usr.UpdateTrashID(ctx, trashID)
		assert.NoError(t, err)

		retrievedUser, err := user.GetUserByUsername(ctx, usr.Username)
		assert.NoError(t, err)
		assert.Equal(t, usr.Username, retrievedUser.Username)
		assert.Equal(t, trashID, retrievedUser.TrashID)
	})
}
