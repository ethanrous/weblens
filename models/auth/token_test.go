package auth_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ethanrous/weblens/models/auth"
)

// BSON keys.
const (
	KeyTokenID   = "_id"
	KeyOwner     = "owner"
	KeyCreatedBy = "createdBy"
	KeyToken     = "token"
)

var sampleTokenCounter = 0

func makeSampleToken() [32]byte {
	token := [32]byte{}
	copy(token[:], "sampletoken"+strconv.Itoa(sampleTokenCounter))

	sampleTokenCounter++

	return token
}

func TestGenerateNewToken(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		token, err := auth.GenerateNewToken(ctx, "nickname", "owner", "createdBy")

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "nickname", token.Nickname)
		assert.Equal(t, "owner", token.Owner)
	})
}

func TestSaveToken(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		token := &auth.Token{
			ID:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}

		err := auth.SaveToken(ctx, token)

		assert.NoError(t, err)
	})
}

func TestGetTokensByUser(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		// Create a mock token
		token := &auth.Token{
			ID:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "towerID",
		}
		err := auth.SaveToken(ctx, token)
		assert.NoError(t, err)

		tokens, err := auth.GetTokensByUser(ctx, "owner")

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Len(t, tokens, 1)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		tokens, err := auth.GetTokensByUser(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})
}

func TestGetTokenByID(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		tokenID := primitive.NewObjectID()
		// Create a mock token
		token := &auth.Token{
			ID:          tokenID,
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := auth.SaveToken(ctx, token)
		assert.NoError(t, err)

		token, err = auth.GetTokenByID(ctx, tokenID)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, tokenID, token.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		token, err := auth.GetTokenByID(ctx, primitive.NewObjectID())

		assert.Nil(t, token)
		assert.Equal(t, auth.ErrTokenNotFound, err)
	})
}

func TestGetToken(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		// Create a mock token
		sampleToken := makeSampleToken()
		token := auth.Token{
			ID:          primitive.NewObjectID(),
			Token:       sampleToken,
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := auth.SaveToken(ctx, &token)
		assert.NoError(t, err)

		token, err = auth.GetToken(ctx, sampleToken)

		assert.NoError(t, err)
		assert.NotNil(t, token)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		tokenBytes := [32]byte{}
		token, err := auth.GetToken(ctx, tokenBytes)

		assert.Error(t, err)
		assert.Equal(t, auth.Token{}, token)
	})
}

func TestGetAllTokensByTowerID(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		towerID := "mockTowerID"
		// Create a mock token
		token := &auth.Token{
			ID:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   towerID,
		}
		err := auth.SaveToken(ctx, token)
		assert.NoError(t, err)

		tokens, err := auth.GetAllTokensByTowerID(ctx, towerID)

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Len(t, tokens, 1)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		tokens, err := auth.GetAllTokensByTowerID(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})
}

func TestDeleteToken(t *testing.T) {
	ctx := db.SetupTestDB(t, auth.TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerID", "towerID") //nolint:revive
		tokenID := primitive.NewObjectID()
		// Create a mock token
		token := &auth.Token{
			ID:          tokenID,
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := auth.SaveToken(ctx, token)
		assert.NoError(t, err)

		err = auth.DeleteToken(ctx, tokenID)

		assert.NoError(t, err)

		// Verify deletion
		token, err = auth.GetTokenByID(ctx, tokenID)
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, auth.ErrTokenNotFound, err)
	})
}
