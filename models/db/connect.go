package db

import (
	"context"
	"time"

	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB is an interface that wraps the mongo.Database type.
type DB interface {
	mongo.Database
}

const maxRetries = 5

// ConnectToMongo establishes a connection to MongoDB with automatic retries and context-aware cleanup.
func ConnectToMongo(ctx context.Context, mongoURI, mongoDbName string) (*mongo.Database, error) {
	l := log.FromContext(ctx)
	l.Debug().CallerSkipFrame(1).Msgf("Connecting to Mongo at %s with name %s ...", mongoURI, mongoDbName)

	clientOptions := options.Client().ApplyURI(mongoURI).SetTimeout(5 * time.Second)

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
			l.Debug().Msg("Context done, stopping MongoDB connection attempts")

			return nil, ctx.Err()
		default:
		}

		l.Warn().Err(err).Msgf("Failed to connect to mongo, retrying in 2s. (%d retries remain)", maxRetries-retries)
		time.Sleep(time.Second)

		retries++
	}

	if err != nil {
		l.Error().Msgf("Failed to connect to database after %d retries", maxRetries)

		return nil, err
	}

	err = context_mod.AddToWg(ctx)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		l.Debug().Msg("Disconnecting from MongoDB...")

		if err := mongoc.Disconnect(context.Background()); err != nil {
			l.Error().Err(err).Msg("Failed to disconnect from MongoDB")
		}

		err := context_mod.WgDone(ctx)
		if err != nil {
			l.Error().Err(err).Msg("Failed to mark WaitGroup done after MongoDB disconnection")
		}
	})

	return mongoc.Database(mongoDbName), nil
}
