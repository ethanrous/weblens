package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/interfaces"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Weblensdb struct {
	mongo *mongo.Database
	redis *redis.Client
}
var mongo_ctx = context.TODO()
//var redis_ctx = context.TODO()
var mongoc *mongo.Client
var mongodb *mongo.Database
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
		mongodb = mongoc.Database("weblens")
	}
	if redisc == nil {
		redisc = redis.NewClient(&redis.Options{
			Addr: util.GetRedisUrl(),
			Password: "",
			DB:		  0,
		})
	}

	return Weblensdb{
		mongo: mongodb,
		redis: redisc,
	}
}

func (db Weblensdb) GetMedia(fileHash string, includeThumbnail bool) (interfaces.Media) {
	conn := db.redis.Get(fileHash)
	val, err := conn.Result()
	if err != nil {

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

func (db Weblensdb) GetMediaByFilepath(filepath string, includeThumbnail bool) (interfaces.Media) {
	filepath = util.GuaranteeRelativePath(filepath)

	filter := bson.D{{Key: "filepath", Value: filepath}}

	var opts *options.FindOneOptions
	if !includeThumbnail {
		opts = options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: util.IntFromBool(includeThumbnail)}})
	}
	findRet := db.mongo.Collection("images").FindOne(mongo_ctx, filter, opts)

	if findRet.Err() != nil {
		return interfaces.Media{}
	}
	var i interfaces.Media
	findRet.Decode(&i)
	return i
}

// Returns ids of all media in directory with depth 1
func (db Weblensdb) GetMediaInDirectory(dirpath string, recursive bool) ([]interfaces.Media) {
	var re string
	if !recursive {
		re = fmt.Sprintf("^%s\\/?[^\\/]+$", util.GuaranteeRelativePath(dirpath))
	} else {
		re = fmt.Sprintf("^%s/?.*$", util.GuaranteeRelativePath(dirpath))
	}

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
		rawMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "mediaType.israw", Value: raw}}}}
		pipeline = append(pipeline, rawMatchStage)
	}

	mediaMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "mediaType.friendlyname", Value: bson.D{{Key: "$ne", Value: "File"}}}}}}
	pipeline = append(pipeline, mediaMatchStage)

	skipStage := bson.D{{Key: "$skip", Value: skip}}
	pipeline = append(pipeline, skipStage)

	if limit != 0 {
		limitStage := bson.D{{Key: "$limit", Value: limit}}
		pipeline = append(pipeline, limitStage)
	}

	opts := options.Aggregate()
	// opts.SetHint("createDate_1")
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
		go db.redisCacheThumbBytes(val)
		res[i].Thumbnail64 = ""
	}

	util.Debug.Println(res)
	return res, len(res) == limit

}

func (db Weblensdb) RedisCacheSet(key string, data string) (error) {
	_, err := db.redis.Set(key, data, time.Duration(600000000000)).Result()
	return err
}

func (db Weblensdb) RedisCacheGet(key string) (string, error) {
	data, err := db.redis.Get(key).Result()
	return data, err
}

func (db Weblensdb) redisCacheThumbBytes(media interfaces.Media) {
	_, err := db.redis.Set(media.FileHash, media, time.Duration(60000000000)).Result()
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) DbAddMedia(m *interfaces.Media) {
	if !m.IsFilledOut(false) {
		util.Error.Panicf("Refusing to write incomplete media to database for file %s", m.Filepath)
	}

	m.Filepath = util.GuaranteeRelativePath(m.Filepath)

	filter := bson.D{{Key: "_id", Value: m.FileHash}}
	set := bson.D{{Key: "$set", Value: *m}}
	_, err := db.mongo.Collection("images").UpdateOne(mongo_ctx, filter, set, options.Update().SetUpsert(true))
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) UpdateMediaByFilepath(filepath string, m interfaces.Media) {
	filepath = util.GuaranteeRelativePath(filepath)
	m.Filepath = util.GuaranteeRelativePath(m.Filepath)

	filter := bson.D{{Key: "filepath", Value: filepath}}
	set := bson.D{{Key: "$set", Value: m}}
	_, err := db.mongo.Collection("images").UpdateOne(mongo_ctx, filter, set)
	util.FailOnError(err, "Failed to update media by filepath")
}

func (db Weblensdb) MoveMedia(oldFilepath string, newM interfaces.Media) {
	oldFilepath = util.GuaranteeRelativePath(oldFilepath)
	newM.Filepath = util.GuaranteeRelativePath(newM.Filepath)

	db.RemoveMediaByFilepath(oldFilepath)
	db.DbAddMedia(&newM)
}

func (db Weblensdb) RemoveMediaByFilepath(filepath string) {
	filter := bson.D{{Key: "filepath", Value: filepath}}
	_, err := db.mongo.Collection("images").DeleteOne(mongo_ctx, filter)
	util.FailOnError(err, "Failed to remove media by filepath")
}

func (db Weblensdb) CreateTrashEntry(originalFilepath, trashPath string) {

	originalFilepath = util.GuaranteeRelativePath(originalFilepath)

	entry := interfaces.TrashEntry{OriginalPath: originalFilepath, TrashPath: trashPath}

	_, err := db.mongo.Collection("trash").InsertOne(mongo_ctx, entry)
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) CheckLogin(username string, password string) bool {
	filter := bson.D{{Key: "username", Value: username}}
	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user interfaces.User
	err := ret.Decode(&user)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (db Weblensdb) AddTokenToUser(username string, token string) {
	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "tokens", Value: token}}}}
	db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
}