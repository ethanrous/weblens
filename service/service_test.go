package service_test

import (
	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/models"
	"go.mongodb.org/mongo-driver/mongo"
)

var mondb *mongo.Database

func init() {
	var err error
	mondb, err = database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
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
