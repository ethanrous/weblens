package database

import (
	"context"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseService struct {
	ctx context.Context

	media       *mongo.Collection
	users       *mongo.Collection
	fileHistory *mongo.Collection
	albums      *mongo.Collection
	shares      *mongo.Collection
	servers     *mongo.Collection
}

func New() types.DatabaseService {
	var uri = util.GetMongoURI()

	clientOptions := options.Client().ApplyURI(uri).SetTimeout(time.Second)
	var err error
	ctx := context.TODO()
	mongoc, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	mongodb := mongoc.Database(util.GetMongoDBName())

	return &databaseService{
		ctx:         ctx,
		media:       mongodb.Collection("media"),
		users:       mongodb.Collection("users"),
		fileHistory: mongodb.Collection("fileHistory"),
		albums:      mongodb.Collection("albums"),
		shares:      mongodb.Collection("shares"),
		servers:     mongodb.Collection("servers"),
	}
}
