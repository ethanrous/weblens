package database

import (
	"context"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseService struct {
	ctx context.Context

	mongoClient *mongo.Client

	media       *mongo.Collection
	users       *mongo.Collection
	fileHistory *mongo.Collection
	albums      *mongo.Collection
	shares      *mongo.Collection
	servers     *mongo.Collection
	trash       *mongo.Collection
	apiKeys     *mongo.Collection
}

const maxRetries = 5

func New(mongoUri, mongoDbName string) types.StoreService {
	clientOptions := options.Client().ApplyURI(mongoUri).SetTimeout(time.Second * 2)
	var err error
	ctx := context.TODO()
	mongoc, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	retries := 0
	for retries < maxRetries {
		err = mongoc.Ping(context.TODO(), nil)
		if err == nil {
			break
		}
		wlog.Warning.Printf("Failed to connect to mongo, trying %d more time(s)", maxRetries-retries)
		time.Sleep(time.Second * 5)
		retries++
	}
	if err != nil {
		wlog.Error.Printf("Failed to connect to database after %d retries", maxRetries)
		panic(err)
	}

	wlog.Debug.Println("Connected to mongo")

	mongodb := mongoc.Database(mongoDbName)

	return &databaseService{
		ctx:         ctx,
		mongoClient: mongoc,
		media:       mongodb.Collection("media"),
		users:       mongodb.Collection("users"),
		fileHistory: mongodb.Collection("fileHistory"),
		albums:      mongodb.Collection("albums"),
		shares:      mongodb.Collection("shares"),
		servers:     mongodb.Collection("servers"),
		trash:       mongodb.Collection("trash"),
		apiKeys:     mongodb.Collection("apiKeys"),
	}
}
