package db

import (
	"context"
	"time"

	context_mod "github.com/ethanrous/weblens/modules/context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB interface {
	mongo.Database
}

const maxRetries = 5

func ConnectToMongo(ctx context.Context, mongoUri, mongoDbName string) (*mongo.Database, error) {
	log := context_mod.ToZ(ctx).Log()
	log.Debug().Msgf("Connecting to Mongo at %s with name %s ...", mongoUri, mongoDbName)

	clientOptions := options.Client().ApplyURI(mongoUri)

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

		select {
		case <-ctx.Done():
			log.Debug().Msg("Context done, stopping MongoDB connection attempts")

			return nil, ctx.Err()
		default:
		}

		log.Warn().Err(err).Msgf("Failed to connect to mongo, retrying in 2s. (%d retries remain)", maxRetries-retries)
		time.Sleep(time.Second)

		retries++
	}

	if err != nil {
		log.Error().Msgf("Failed to connect to database after %d retries", maxRetries)

		return nil, err
	}

	context.AfterFunc(ctx, func() {
		log.Debug().Msg("Disconnecting from MongoDB...")

		if err := mongoc.Disconnect(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to disconnect from MongoDB")
		}
	})

	return mongoc.Database(mongoDbName), nil
}
