
package cryptography_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	t.Run("generates random int within limit", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			result, err := cryptography.RandomInt(10)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, result, int64(0))
			assert.Less(t, result, int64(10))
		}
	})

	t.Run("generates random int with large limit", func(t *testing.T) {
		result, err := cryptography.RandomInt(1000000)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result, int64(0))
		assert.Less(t, result, int64(1000000))
	})

	t.Run("generates different values", func(t *testing.T) {
		seen := make(map[int64]bool)
		for i := 0; i < 100; i++ {
			result, err := cryptography.RandomInt(1000)
			require.NoError(t, err)
			seen[result] = true
		}
		// Should have generated multiple different values
		assert.Greater(t, len(seen), 1)
	})
}

func TestRandomString(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		result, err := cryptography.RandomString(10)
		require.NoError(t, err)
		assert.Equal(t, 10, len(result))
	})

	t.Run("generates string with only alphanumeric chars", func(t *testing.T) {
		result, err := cryptography.RandomString(100)
		require.NoError(t, err)
		for _, c := range result {
			isLower := c >= 'a' && c <= 'z'
			isUpper := c >= 'A' && c <= 'Z'
			isDigit := c >= '0' && c <= '9'
			assert.True(t, isLower || isUpper || isDigit, "unexpected char: %c", c)
		}
	})

	t.Run("generates zero length string", func(t *testing.T) {
		result, err := cryptography.RandomString(0)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("generates different strings", func(t *testing.T) {
		s1, err := cryptography.RandomString(32)
		require.NoError(t, err)
		s2, err := cryptography.RandomString(32)
		require.NoError(t, err)
		assert.NotEqual(t, s1, s2)
	})
}

func TestRandomBytes(t *testing.T) {
	t.Run("generates bytes of correct length", func(t *testing.T) {
		result, err := cryptography.RandomBytes(16)
		require.NoError(t, err)
		assert.Equal(t, 16, len(result))
	})

	t.Run("generates zero length bytes", func(t *testing.T) {
		result, err := cryptography.RandomBytes(0)
		require.NoError(t, err)
		assert.Equal(t, 0, len(result))
	})

	t.Run("generates different byte sequences", func(t *testing.T) {
		b1, err := cryptography.RandomBytes(32)
		require.NoError(t, err)
		b2, err := cryptography.RandomBytes(32)
		require.NoError(t, err)
		assert.NotEqual(t, b1, b2)
	})
}

func TestHashString(t *testing.T) {
	t.Run("hashes string to base64", func(t *testing.T) {
		result := cryptography.HashString("hello")
		assert.NotEmpty(t, result)
		// SHA-256 produces 32 bytes, base64 encodes to ~43 chars
		assert.Greater(t, len(result), 40)
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		h1 := cryptography.HashString("test")
		h2 := cryptography.HashString("test")
		assert.Equal(t, h1, h2)
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := cryptography.HashString("test1")
		h2 := cryptography.HashString("test2")
		assert.NotEqual(t, h1, h2)
	})

	t.Run("empty string can be hashed", func(t *testing.T) {
		result := cryptography.HashString("")
		assert.NotEmpty(t, result)
	})
}

func TestHashUserPassword(t *testing.T) {
	// Use minimal bcrypt cost for faster tests
	ctx := context.WithValue(context.Background(), cryptography.BcryptDifficultyCtxKey, 4)

	t.Run("hashes password", func(t *testing.T) {
		hash, err := cryptography.HashUserPassword(ctx, "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, "password123", hash)
	})

	t.Run("same password produces different hashes", func(t *testing.T) {
		h1, err := cryptography.HashUserPassword(ctx, "password")
		require.NoError(t, err)
		h2, err := cryptography.HashUserPassword(ctx, "password")
		require.NoError(t, err)
		// bcrypt uses random salt, so hashes should differ
		assert.NotEqual(t, h1, h2)
	})
}

func TestVerifyUserPassword(t *testing.T) {
	// Use minimal bcrypt cost for faster tests
	ctx := context.WithValue(context.Background(), cryptography.BcryptDifficultyCtxKey, 4)

	t.Run("verifies correct password", func(t *testing.T) {
		password := "correctPassword123"
		hash, err := cryptography.HashUserPassword(ctx, password)
		require.NoError(t, err)

		err = cryptography.VerifyUserPassword(password, hash)
		assert.NoError(t, err)
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		password := "correctPassword"
		hash, err := cryptography.HashUserPassword(ctx, password)
		require.NoError(t, err)

		err = cryptography.VerifyUserPassword("wrongPassword", hash)
		assert.Error(t, err)
	})
}

func TestGenerateJWT(t *testing.T) {
	t.Run("generates valid JWT", func(t *testing.T) {
		token, expires, err := cryptography.GenerateJWT("testuser")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, expires.After(time.Now()))
	})

	t.Run("expiration is in the future", func(t *testing.T) {
		_, expires, err := cryptography.GenerateJWT("testuser")
		require.NoError(t, err)
		// Should expire approximately 7 days from now
		expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
		assert.WithinDuration(t, expectedExpiry, expires, time.Minute)
	})

	t.Run("different users get different tokens", func(t *testing.T) {
		t1, _, err := cryptography.GenerateJWT("user1")
		require.NoError(t, err)
		t2, _, err := cryptography.GenerateJWT("user2")
		require.NoError(t, err)
		assert.NotEqual(t, t1, t2)
	})
}

func TestGetUsernameFromToken(t *testing.T) {
	t.Run("extracts username from valid token", func(t *testing.T) {
		username := "testuser"
		token, _, err := cryptography.GenerateJWT(username)
		require.NoError(t, err)

		extractedUsername, err := cryptography.GetUsernameFromToken(token)
		require.NoError(t, err)
		assert.Equal(t, username, extractedUsername)
	})

	t.Run("returns error for empty token", func(t *testing.T) {
		_, err := cryptography.GetUsernameFromToken("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no jwt provided")
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		_, err := cryptography.GetUsernameFromToken("invalid.token.here")
		assert.Error(t, err)
	})

	t.Run("returns error for malformed token", func(t *testing.T) {
		_, err := cryptography.GetUsernameFromToken("notajwt")
		assert.Error(t, err)
	})
}

func TestRoundTripJWT(t *testing.T) {
	t.Run("can generate and extract username", func(t *testing.T) {
		testCases := []string{
			"simple",
			"user@example.com",
			"user-with-dashes",
			"user_with_underscores",
			"MixedCase123",
		}

		for _, username := range testCases {
			t.Run(username, func(t *testing.T) {
				token, _, err := cryptography.GenerateJWT(username)
				require.NoError(t, err)

				extracted, err := cryptography.GetUsernameFromToken(token)
				require.NoError(t, err)
				assert.Equal(t, username, extracted)
			})
		}
	})
}
