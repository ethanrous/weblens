// Package auth provides authentication token management for the Weblens system.
package auth

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenCollectionKey is the MongoDB collection name for storing authentication tokens.
const TokenCollectionKey = "tokens"

// ErrTokenNotFound is returned when a requested token does not exist in the database.
var ErrTokenNotFound = errors.New("no token found")

// Token represents an authentication token with metadata about its creation, ownership, and usage.
type Token struct {
	CreatedTime time.Time          `bson:"createdTime"`
	LastUsed    time.Time          `bson:"lastUsed"`
	Nickname    string             `bson:"nickname"`
	Owner       string             `bson:"owner"`
	RemoteUsing string             `bson:"remoteUsing"`
	CreatedBy   string             `bson:"createdBy"`
	Token       [32]byte           `bson:"token"`
	ID          primitive.ObjectID `bson:"_id"`
}

// GenerateNewToken creates and saves a new authentication token with the specified nickname, owner, and creator.
func GenerateNewToken(ctx context.Context, nickname, owner, createdBy string) (*Token, error) {
	tok, err := crypto.RandomBytes(32)
	if err != nil {
		return nil, err
	}

	tokenBytes := [32]byte{}
	copy(tokenBytes[:], tok)

	token := &Token{
		CreatedTime: time.Now(),
		LastUsed:    time.Now(),
		Nickname:    nickname,
		Owner:       owner,
		CreatedBy:   createdBy,
		ID:          primitive.NewObjectID(),
		Token:       tokenBytes,
	}

	col, err := db.GetCollection[Token](ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	_, err = col.InsertOne(ctx, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// SaveToken persists an authentication token to the database.
func SaveToken(ctx context.Context, token *Token) error {
	if token.Token == [32]byte{} {
		return errors.New("token is empty")
	}

	col, err := db.GetCollection[any](ctx, TokenCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

// GetTokensByUser retrieves all tokens owned by a specific user that were created by the local tower.
func GetTokensByUser(ctx context.Context, username string) ([]*Token, error) {
	col, err := db.GetCollection[any](ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	localTowerID := ctx.Value("towerID").(string)

	cursor, err := col.Find(ctx, bson.M{"owner": username, "createdBy": localTowerID})
	if err != nil {
		return nil, err
	}

	var tokens []*Token

	for cursor.Next(ctx) {
		var token Token
		if err := cursor.Decode(&token); err != nil {
			return nil, err
		}

		tokens = append(tokens, &token)
	}

	return tokens, nil
}

// GetTokenByID retrieves a token by its database ID.
func GetTokenByID(ctx context.Context, tokenID primitive.ObjectID) (token *Token, err error) {
	col, err := db.GetCollection[*Token](ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	token, err = col.FindOneAs(ctx, bson.M{"_id": tokenID})
	if db.IsNotFound(err) {
		return nil, ErrTokenNotFound
	}

	return
}

// GetToken retrieves a token by its cryptographic token value.
func GetToken(ctx context.Context, tokenBytes [32]byte) (Token, error) {
	col, err := db.GetCollection[Token](ctx, TokenCollectionKey)
	if err != nil {
		return Token{}, err
	}

	token, err := col.FindOneAs(ctx, bson.M{"token": tokenBytes})
	if err != nil {
		return token, db.WrapError(err, "failed to find token")
	}

	return token, nil
}

// GetAllTokensByTowerID retrieves all tokens that were created by a specific tower.
func GetAllTokensByTowerID(ctx context.Context, towerID string) (tokens []*Token, err error) {
	col, err := db.GetCollection[any](ctx, TokenCollectionKey)
	if err != nil {
		return
	}

	res, err := col.Find(ctx, bson.M{"createdBy": towerID})
	if err != nil {
		return
	}

	err = res.All(ctx, &tokens)
	if err != nil {
		return
	}

	return
}

// DeleteToken removes a token from the database by its ID.
func DeleteToken(ctx context.Context, tokenID primitive.ObjectID) error {
	col, err := db.GetCollection[any](ctx, TokenCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteOne(ctx, bson.M{"_id": tokenID})
	if err != nil {
		return err
	}

	return nil
}
