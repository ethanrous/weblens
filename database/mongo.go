package database

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbCollectionName string

const (
	InstanceCollectionKey    DbCollectionName = "servers"
	ApiKeysCollectionKey     DbCollectionName = "apiKeys"
	UsersCollectionKey       DbCollectionName = "users"
	AlbumsCollectionKey      DbCollectionName = "albums"
	SharesCollectionKey      DbCollectionName = "shares"
	FileHistoryCollectionKey DbCollectionName = "fileHistory"
	FolderMediaCollectionKey DbCollectionName = "folderMedia"
	MediaCollectionKey       DbCollectionName = "media"
)

const maxRetries = 5

func ConnectToMongo(mongoUri, mongoDbName string, logger *zerolog.Logger) (*mongo.Database, error) {
	logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Connecting to Mongo at %s with name %s ...", mongoUri, mongoDbName) })
	clientOptions := options.Client().ApplyURI(mongoUri).SetTimeout(time.Second * 5)
	var err error
	mongoc, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	retries := 0
	for retries < maxRetries {
		err = mongoc.Ping(context.Background(), nil)
		if err == nil {
			break
		}
		logger.Warn().Msgf("Failed to connect to mongo, trying %d more time(s)", maxRetries-retries)
		time.Sleep(time.Second * 1)
		retries++
	}
	if err != nil {
		logger.Error().Msgf("Failed to connect to database after %d retries", maxRetries)
		return nil, err
	}

	return mongoc.Database(mongoDbName), nil
}

type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (
		*mongo.InsertOneResult, error,
	)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error)
	UpdateOne(
		ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}
