package db

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB interface {
	mongo.Database
}

const maxRetries = 5

func ConnectToMongo(ctx context.Context, mongoUri, mongoDbName string) (*mongo.Database, error) {
	log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Connecting to Mongo at %s with name %s ...", mongoUri, mongoDbName) })

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoUri).SetTimeout(time.Second * 10)
	var err error
	mongoc, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	retries := 0
	for retries < maxRetries {
		err = mongoc.Ping(ctx, nil)
		if err == nil {
			break
		}
		log.Warn().Msgf("Failed to connect to mongo, retrying in 2s. (%d retries remain)", maxRetries-retries)
		time.Sleep(time.Second * 2)
		retries++
	}
	if err != nil {
		log.Error().Msgf("Failed to connect to database after %d retries", maxRetries)
		return nil, err
	}

	return mongoc.Database(mongoDbName), nil
}
