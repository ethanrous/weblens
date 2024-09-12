package database

import (
	"context"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxRetries = 5

func ConnectToMongo(mongoUri, mongoDbName string) (*mongo.Database, error) {
	log.Debug.Printf("Connecting to Mongo at %s", mongoUri)
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
		log.Warning.Printf("Failed to connect to mongo, trying %d more time(s)", maxRetries-retries)
		time.Sleep(time.Second * 1)
		retries++
	}
	if err != nil {
		log.Error.Printf("Failed to connect to database after %d retries", maxRetries)
		return nil, err
	}

	log.Debug.Println("Connected to mongo")
	log.Debug.Printf("Using Mongo database %s", mongoDbName)

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
