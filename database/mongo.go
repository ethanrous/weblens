package database

import (
	"context"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxRetries = 5

func ConnectToMongo(mongoUri, mongoDbName string) *mongo.Database {
	clientOptions := options.Client().ApplyURI(mongoUri).SetTimeout(time.Second * 5)
	var err error
	mongoc, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	log.Debug.Println("Connected to mongo")

	return mongoc.Database(mongoDbName)
}
