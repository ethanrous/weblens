package service_test

import (
	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/models"
	"go.mongodb.org/mongo-driver/mongo"
)

var mondb *mongo.Database

func init() {
	var err error
	mondb, err = database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}))
	if err != nil {
		panic(err)
	}

	marshMap := map[string]models.MediaType{}
	err = env.ReadTypesConfig(&marshMap)
	if err != nil {
		panic(err)
	}
	typeService = models.NewTypeService(marshMap)
}
