package auth

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const TokenCollectionKey = "tokens"

var ErrTokenNotFound = errors.New("no token found")

type Token struct {
	CreatedTime time.Time          `bson:"createdTime"`
	LastUsed    time.Time          `bson:"lastUsed"`
	Nickname    string             `bson:"nickname"`
	Owner       string             `bson:"owner"`
	RemoteUsing string             `bson:"remoteUsing"`
	CreatedBy   string             `bson:"createdBy"`
	Token       [32]byte           `bson:"token"`
	Id          primitive.ObjectID `bson:"_id"`
}

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
		Id:          primitive.NewObjectID(),
		Token:       tokenBytes,
	}

	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	_, err = col.InsertOne(ctx, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func SaveToken(ctx context.Context, token *Token) error {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func GetTokensByUser(ctx context.Context, username string) ([]*Token, error) {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	localTowerId := ctx.Value("towerId").(string)

	cursor, err := col.Find(ctx, bson.M{"owner": username, "createdBy": localTowerId})
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

func GetTokenById(ctx context.Context, tokenId primitive.ObjectID) (token *Token, err error) {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return nil, err
	}

	token = &Token{}

	err = col.FindOne(ctx, bson.M{"_id": tokenId}).Decode(token)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrTokenNotFound
	}

	return
}

func GetToken(ctx context.Context, tokenBytes [32]byte) (token Token, err error) {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return
	}

	err = col.FindOne(ctx, bson.M{"token": tokenBytes}).Decode(&token)
	if err != nil {
		return token, db.WrapError(err, "failed to find token")
	}

	return
}

func GetAllTokensByTowerId(ctx context.Context, towerId string) (tokens []*Token, err error) {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return
	}

	res, err := col.Find(ctx, bson.M{"towerId": towerId})
	if err != nil {
		return
	}

	err = res.All(ctx, &tokens)
	if err != nil {
		return
	}

	return
}

func DeleteToken(ctx context.Context, tokenId primitive.ObjectID) error {
	col, err := db.GetCollection(ctx, TokenCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteOne(ctx, bson.M{"_id": tokenId})
	if err != nil {
		return err
	}

	return nil
}
