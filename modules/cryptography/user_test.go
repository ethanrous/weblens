package cryptography_test

import (
	"testing"

	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTKeyFromEnv(t *testing.T) {
	t.Run("uses env var when set", func(t *testing.T) {
		t.Setenv("WEBLENS_JWT_SECRET", "test-secret-key-that-is-long-enough-32chars!")
		cryptography.ReloadJWTKey()

		token, _, err := cryptography.GenerateJWT("testuser")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		username, err := cryptography.GetUsernameFromToken(token)
		require.NoError(t, err)
		assert.Equal(t, "testuser", username)
	})

	t.Run("generates random key when env var unset", func(t *testing.T) {
		t.Setenv("WEBLENS_JWT_SECRET", "")
		cryptography.ReloadJWTKey()

		token, _, err := cryptography.GenerateJWT("testuser")
		require.NoError(t, err)

		username, err := cryptography.GetUsernameFromToken(token)
		require.NoError(t, err)
		assert.Equal(t, "testuser", username)
	})

	t.Run("tokens from different keys are incompatible", func(t *testing.T) {
		t.Setenv("WEBLENS_JWT_SECRET", "key-one-aaaaaaaaaaaaaaaaaaaaaaaaa")
		cryptography.ReloadJWTKey()

		token1, _, err := cryptography.GenerateJWT("testuser")
		require.NoError(t, err)

		t.Setenv("WEBLENS_JWT_SECRET", "key-two-bbbbbbbbbbbbbbbbbbbbbbbbb")
		cryptography.ReloadJWTKey()

		_, err = cryptography.GetUsernameFromToken(token1)
		assert.Error(t, err, "token signed with different key should be rejected")
	})
}

func TestValidateFilename(t *testing.T) {
	t.Run("rejects invalid filenames", func(t *testing.T) {
		invalidNames := []struct {
			name  string
			input string
		}{
			{"path traversal parent", ".."},
			{"path traversal with prefix", "../etc/passwd"},
			{"path traversal embedded", "foo/../bar"},
			{"forward slash", "foo/bar"},
			{"backslash", "foo\\bar"},
			{"dot current dir", "."},
			{"empty string", ""},
		}

		for _, tt := range invalidNames {
			t.Run(tt.name, func(t *testing.T) {
				assert.Error(t, cryptography.ValidateFilename(tt.input))
			})
		}
	})

	t.Run("accepts valid filenames", func(t *testing.T) {
		validNames := []string{
			"photo.jpg",
			"my document.pdf",
			".hidden",
			"file-name_v2.tar.gz",
		}

		for _, name := range validNames {
			t.Run(name, func(t *testing.T) {
				assert.NoError(t, cryptography.ValidateFilename(name))
			})
		}
	})
}
