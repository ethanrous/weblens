package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/interfaces"
	util "github.com/ethrousseau/weblens/api/utils"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Weblensdb struct {
	mongo *mongo.Database
	redis *redis.Client
}
var mongo_ctx = context.TODO()
var redis_ctx = context.TODO()
var mongoc *mongo.Client
var redisc *redis.Client

func New() (Weblensdb) {
	if mongoc == nil {
		var uri = util.GetMongoURI()
		clientOptions := options.Client().ApplyURI(uri)
		var err error
		mongoc, err = mongo.Connect(mongo_ctx, clientOptions)
		if err != nil {
			panic(err)
		}
	}
	if redisc == nil {
		redisc = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			Password: "",
			DB:		  0,
		})
	}

	return Weblensdb{
		mongo: mongoc.Database("weblens"),
		redis: redisc,
	}
}

func (db Weblensdb) GetMedia(fileHash string, includeThumbnail bool) (interfaces.Media) {
	conn := db.redis.Get(fileHash)
	val, err := conn.Result()
	if err != nil {
		//util.Debug.Print("Redis cache miss")

		var opts *options.FindOneOptions

		if !includeThumbnail {
			opts = options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: util.IntFromBool(includeThumbnail)}})
		}

		filter := bson.D{{Key: "_id", Value: fileHash}}
		findRet := db.mongo.Collection("images").FindOne(mongo_ctx, filter, opts)

		var i interfaces.Media
		findRet.Decode(&i)
		return i

	} else {
		//fmt.Println("Redis cache hit")
		var i interfaces.Media
		json.Unmarshal([]byte(val), &i)
		return i
	}
}

// Returns image if found and bool for if image exists in db
func (db Weblensdb) GetMediaByFilepath(filepath string) (interfaces.Media, bool) {
	filter := bson.D{{Key: "filepath", Value: filepath}}
	opts := options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: 0}})
	findRet := db.mongo.Collection("images").FindOne(mongo_ctx, filter, opts)

	if findRet.Err() != nil {
		return interfaces.Media{}, false
	}
	var i interfaces.Media
	findRet.Decode(&i)
	return i, true
}

// Returns ids of all media in directory with depth 1
func (db Weblensdb) GetMediaInDirectory(filepath string) ([]interfaces.Media) {
	re := fmt.Sprintf("^%s\\/?[^\\/]+$", filepath)

	filter := bson.M{"filepath": bson.M{"$regex": re, "$options": "i"}}
	opts := options.Find().SetProjection(bson.D{{Key: "thumbnail", Value: 0}})

	findRet, err := db.mongo.Collection("images").Find(mongo_ctx, filter, opts)

	if err != nil {
		panic(err)
	}

	var i []interfaces.Media
	findRet.All(mongo_ctx, &i)
	return i
}

func (db Weblensdb) GetPagedMedia(sort string, skip, limit int, raw, thumbnails bool) ([]interfaces.Media, bool) {
	pipeline := mongo.Pipeline{}

	if !thumbnails {
		unsetStage := bson.D{{Key: "$unset", Value: bson.A{"thumbnail"}}}
		pipeline = append(pipeline, unsetStage)
	}


	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: -1}}}}
	pipeline = append(pipeline, sortStage)

	if !raw {
		matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "mediaType.israw", Value: raw}}}}
		pipeline = append(pipeline, matchStage)
	}

	skipStage := bson.D{{Key: "$skip", Value: skip}}
	pipeline = append(pipeline, skipStage)

	if limit != 0 {
		limitStage := bson.D{{Key: "$limit", Value: limit}}
		pipeline = append(pipeline, limitStage)
	}

	opts := options.Aggregate()
	opts.SetHint("createDate_1")
	cursor, err := db.mongo.Collection("images").Aggregate(mongo_ctx, pipeline, opts)
	if err != nil {
		panic(err)
	}

	var res []interfaces.Media
	err = cursor.All(mongo_ctx, &res)
	if err != nil {
		panic(err)
	}

	for i, val := range res {
		db.redisCacheThumbBytes(val)
		res[i].Thumbnail64 = ""
	}
	return res, len(res) == limit

}

func (db Weblensdb) redisCacheThumbBytes(media interfaces.Media) {
	_, err := db.redis.Set(media.FileHash, media, time.Duration(60000000000)).Result()
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) DbAddMedia(media *interfaces.Media) {
	filter := bson.D{{Key: "_id", Value: media.FileHash}}
	set := bson.D{{Key: "$set", Value: *media}}
	_, err := db.mongo.Collection("images").UpdateOne(mongo_ctx, filter, set, options.Update().SetUpsert(true))
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) RemoveMediaByFilepath(filepath string) {
	filter := bson.D{{Key: "filepath", Value: filepath}}
	_, err := db.mongo.Collection("images").DeleteOne(mongo_ctx, filter)
	if err != nil {
		panic(err)
	}
}