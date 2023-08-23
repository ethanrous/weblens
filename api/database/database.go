package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ethrousseau/weblens/api/interfaces"
)

const uri = "mongodb://localhost:27017"

var db *mongo.Database
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	db = client.Database("weblens")
}

func getImage(ID primitive.ObjectID) ([]byte) {
	image := &interfaces.Media{
		ID: ID,
	}
	findRet := db.Collection("images").FindOne(ctx, image)
	var ret []byte
	findRet.Decode(ret)
	return ret
}

// Returns image if found and bool for if image exists in db
func getImageByFilename(filename string) (interfaces.Media, bool) {
	filter := bson.D{{Key: "filename", Value: filename}}
	findRet := db.Collection("images").FindOne(ctx, filter)

	if findRet.Err() != nil {
		return interfaces.Media{}, false
	}
	var i interfaces.Media
	findRet.Decode(&i)
	return i, true
}

/*
func getImageThumb(filehash string) (string) {
	filter := bson.D{{Key: "fileHash", Value: filehash}}
	findRet := db.Collection("thumbnails").FindOne(ctx, filter)

	var t Thumbnail
	findRet.Decode(&t)
	return t.Thumbnail64
}
*/

func getAllImages() ([]interfaces.Media) {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "createDate", Value: -1}})
	findRet, err := db.Collection("images").Find(ctx, filter, opts)
	if err != nil {
		panic(err)
	}

	var ret []interfaces.Media
	err = findRet.All(ctx, &ret)

	if err != nil {
		panic(err)
	}

	return ret

}

func addImage(image interfaces.Media) () {
	image.ID = primitive.NewObjectID()

	_, err := db.Collection("images").InsertOne(ctx, image)
	if err != nil {
		panic(err)
	}

}