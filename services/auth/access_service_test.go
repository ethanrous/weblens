package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCanUserAccessFile_OwnerAccess(t *testing.T) {
	ctx := context.Background()

	// Create a test user
	testUser := &user_model.User{
		Username: "testuser",
	}

	// Create a file owned by testuser
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:    filepath,
		MemOnly: true,
	})

	// Owner should have full access without a share
	perms, err := auth.CanUserAccessFile(ctx, testUser, file, nil)
	require.NoError(t, err)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanDownload)
	assert.True(t, perms.CanEdit)
	assert.True(t, perms.CanDelete)
}

func TestCanUserAccessFile_NilUser(t *testing.T) {
	ctx := context.Background()

	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepath,
		MemOnly:    true,
		GenerateID: true,
	})

	// Nil user should return ErrMustAuthenticate, not panic
	_, err := auth.CanUserAccessFile(ctx, nil, file, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrMustAuthenticate)
}

func TestCanUserAccessFile_NonOwnerNoShare(t *testing.T) {
	ctx := context.Background()

	// Create users
	fileOwner := &user_model.User{Username: "fileowner"}
	otherUser := &user_model.User{Username: "otheruser"}

	// Create a file owned by fileowner
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "fileowner/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:    filepath,
		MemOnly: true,
	})
	_ = fileOwner // Used to indicate ownership via path

	// Other user without share should not have access
	_, err := auth.CanUserAccessFile(ctx, otherUser, file, nil)
	assert.Error(t, err)
}

func TestCanUserAccessFile_PublicUser_PublicShare(t *testing.T) {
	ctx := context.Background()

	// Create a public user
	publicUser := &user_model.User{Username: user_model.PublicUserName}

	// Create a file
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepath,
		MemOnly:    true,
		GenerateID: true,
	})

	// Create a public share for this file
	share := &share_model.FileShare{
		ShareID: primitive.NewObjectID(),
		FileID:  file.ID(),
		Public:  true,
		Enabled: true,
	}

	// Public user should be able to access public share
	perms, err := auth.CanUserAccessFile(ctx, publicUser, file, share)
	require.NoError(t, err)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanDownload)
}

func TestCanUserAccessFile_PublicUser_PrivateShare(t *testing.T) {
	ctx := context.Background()

	// Create a public user
	publicUser := &user_model.User{Username: user_model.PublicUserName}

	// Create a file
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepath,
		MemOnly:    true,
		GenerateID: true,
	})

	// Create a private share for this file
	share := &share_model.FileShare{
		ShareID: primitive.NewObjectID(),
		FileID:  file.ID(),
		Public:  false,
		Enabled: true,
	}

	// Public user should not be able to access private share
	_, err := auth.CanUserAccessFile(ctx, publicUser, file, share)
	assert.Error(t, err)
}

func TestCanUserAccessFile_SystemUser(t *testing.T) {
	ctx := context.Background()

	// Create the WEBLENS system user
	systemUser := &user_model.User{
		Username:  "WEBLENS",
		UserPerms: user_model.UserPermissionSystem,
	}

	// Create a file
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "anyuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepath,
		MemOnly:    true,
		GenerateID: true,
	})

	// Create any share (doesn't matter)
	share := &share_model.FileShare{
		ShareID: primitive.NewObjectID(),
		FileID:  file.ID(),
		Public:  true,
		Enabled: true,
	}

	// System user should have full access
	perms, err := auth.CanUserAccessFile(ctx, systemUser, file, share)
	require.NoError(t, err)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanDownload)
	assert.True(t, perms.CanEdit)
	assert.True(t, perms.CanDelete)
}

func TestCanUserModifyShare(t *testing.T) {
	t.Run("owner can modify share", func(t *testing.T) {
		user := &user_model.User{Username: "shareowner"}
		share := share_model.FileShare{
			Owner: "shareowner",
		}

		assert.True(t, auth.CanUserModifyShare(user, share))
	})

	t.Run("non-owner cannot modify share", func(t *testing.T) {
		user := &user_model.User{Username: "otheruser"}
		share := share_model.FileShare{
			Owner: "shareowner",
		}

		assert.False(t, auth.CanUserModifyShare(user, share))
	})

	t.Run("admin cannot modify share they do not own", func(t *testing.T) {
		admin := &user_model.User{Username: "admin", UserPerms: user_model.UserPermissionAdmin}
		share := share_model.FileShare{
			Owner: "shareowner",
		}

		assert.False(t, auth.CanUserModifyShare(admin, share))
	})
}

func TestGenerateJWTCookie(t *testing.T) {
	t.Run("generates valid cookie string with security flags", func(t *testing.T) {
		user := &user_model.User{Username: "testuser"}

		cookie, err := auth.GenerateJWTCookie(user)
		require.NoError(t, err)
		assert.Contains(t, cookie, "weblens-session-token=")
		assert.Contains(t, cookie, "Path=/")
		assert.Contains(t, cookie, "Expires=")
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Secure")
		assert.Contains(t, cookie, "SameSite=Lax")
	})
}

func TestGenerateUserCookie(t *testing.T) {
	t.Run("generates user cookie with security flags", func(t *testing.T) {
		user := &user_model.User{Username: "testuser"}

		cookie := auth.GenerateUserCookie(user)
		assert.Contains(t, cookie, "weblens-user-name=testuser")
		assert.Contains(t, cookie, "Path=/")
		assert.Contains(t, cookie, "Expires=")
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Secure")
		assert.Contains(t, cookie, "SameSite=Lax")
	})
}

func TestGetUserFromAuthHeader_InvalidFormat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		header string
	}{
		{"empty header", ""},
		{"no bearer prefix", "sometoken"},
		{"short header", "Bear"},
		{"wrong prefix", "Basic sometoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := auth.GetUserFromAuthHeader(ctx, tt.header)
			assert.Error(t, err)
		})
	}
}

func TestCanUserAccessFile_DisabledShare(t *testing.T) {
	ctx := context.Background()

	// Create users
	otherUser := &user_model.User{Username: "otheruser"}

	// Create a file
	filepath := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/image.jpg")
	file := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepath,
		MemOnly:    true,
		GenerateID: true,
	})

	// Create a disabled share
	share := &share_model.FileShare{
		ShareID: primitive.NewObjectID(),
		FileID:  file.ID(),
		Public:  true,
		Enabled: false, // Disabled
	}

	// User should not be able to access disabled share
	_, err := auth.CanUserAccessFile(ctx, otherUser, file, share)
	assert.Error(t, err)
}

func TestCanUserAccessFile_WrongFileForShare(t *testing.T) {
	ctx := context.Background()

	otherUser := &user_model.User{Username: "otheruser"}

	// Create file A
	filepathA := file_system.BuildFilePath(file_model.UsersTreeKey, "testuser/photos/imageA.jpg")
	fileA := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       filepathA,
		MemOnly:    true,
		GenerateID: true,
	})

	// Create a share for a different file (B)
	share := &share_model.FileShare{
		ShareID: primitive.NewObjectID(),
		FileID:  "different-file-id",
		Public:  true,
		Enabled: true,
	}

	// User should not be able to access file A with share for file B
	_, err := auth.CanUserAccessFile(ctx, otherUser, fileA, share)
	assert.Error(t, err)
}

func TestCookieExpiration(t *testing.T) {
	t.Run("user cookie expires in the future", func(t *testing.T) {
		user := &user_model.User{Username: "testuser"}
		cookie := auth.GenerateUserCookie(user)

		// Parse the Expires value from the cookie
		assert.Contains(t, cookie, "Expires=")

		parts := strings.Split(cookie, "Expires=")
		require.Len(t, parts, 2, "cookie should contain Expires=")

		expStr := strings.Split(parts[1], ";")[0]
		expTime, err := time.Parse(time.RFC1123, expStr)
		require.NoError(t, err, "cookie Expires value should parse as RFC1123")
		assert.True(t, expTime.After(time.Now().Add(time.Hour*24*6)), "cookie should expire at least 6 days from now")
	})
}
