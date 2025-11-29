package auth_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	. "github.com/ethanrous/weblens/models/auth"
)

// BSON keys.
const (
	KeyTokenId   = "_id"
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
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		token, err := GenerateNewToken(ctx, "nickname", "owner", "createdBy")

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "nickname", token.Nickname)
		assert.Equal(t, "owner", token.Owner)
	})
}

func TestSaveToken(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		token := &Token{
			Id:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}

		err := SaveToken(ctx, token)

		assert.NoError(t, err)
	})
}

func TestGetTokensByUser(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		// Create a mock token
		token := &Token{
			Id:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "towerId",
		}
		err := SaveToken(ctx, token)
		assert.NoError(t, err)

		tokens, err := GetTokensByUser(ctx, "owner")

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Len(t, tokens, 1)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		tokens, err := GetTokensByUser(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})
}

func TestGetTokenById(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		tokenId := primitive.NewObjectID()
		// Create a mock token
		token := &Token{
			Id:          tokenId,
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := SaveToken(ctx, token)
		assert.NoError(t, err)

		token, err = GetTokenById(ctx, tokenId)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, tokenId, token.Id)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		token, err := GetTokenById(ctx, primitive.NewObjectID())

		assert.Nil(t, token)
		assert.Equal(t, ErrTokenNotFound, err)
	})
}

func TestGetToken(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		// Create a mock token
		sampleToken := makeSampleToken()
		token := Token{
			Id:          primitive.NewObjectID(),
			Token:       sampleToken,
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := SaveToken(ctx, &token)
		assert.NoError(t, err)

		token, err = GetToken(ctx, sampleToken)

		assert.NoError(t, err)
		assert.NotNil(t, token)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		tokenBytes := [32]byte{}
		token, err := GetToken(ctx, tokenBytes)

		assert.Error(t, err)
		assert.Equal(t, Token{}, token)
	})
}

func TestGetAllTokensByTowerId(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("found", func(t *testing.T) {
		towerId := "mockTowerId"
		// Create a mock token
		token := &Token{
			Id:          primitive.NewObjectID(),
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   towerId,
		}
		err := SaveToken(ctx, token)
		assert.NoError(t, err)

		tokens, err := GetAllTokensByTowerId(ctx, towerId)

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Len(t, tokens, 1)
	})

	t.Run("not found", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		tokens, err := GetAllTokensByTowerId(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})
}

func TestDeleteToken(t *testing.T) {
	ctx := db.SetupTestDB(t, TokenCollectionKey)

	t.Run("success", func(t *testing.T) {
		ctx = context.WithValue(ctx, "towerId", "towerId")
		tokenId := primitive.NewObjectID()
		// Create a mock token
		token := &Token{
			Id:          tokenId,
			Token:       makeSampleToken(),
			CreatedTime: time.Now(),
			LastUsed:    time.Now(),
			Nickname:    "nickname",
			Owner:       "owner",
			CreatedBy:   "createdBy",
		}
		err := SaveToken(ctx, token)
		assert.NoError(t, err)

		err = DeleteToken(ctx, tokenId)

		assert.NoError(t, err)

		// Verify deletion
		token, err = GetTokenById(ctx, tokenId)
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, ErrTokenNotFound, err)
	})
}
