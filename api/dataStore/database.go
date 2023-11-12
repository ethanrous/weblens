package dataStore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Weblensdb struct {
	mongo 		*mongo.Database
	redis 		*redis.Client
	accessor 	string
}
var mongo_ctx = context.TODO()
//var redis_ctx = context.TODO()
var mongoc *mongo.Client
var mongodb *mongo.Database
var redisc *redis.Client

func NewDB(username string) (Weblensdb) {
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
	if redisc == nil && util.ShouldUseRedis() {
		redisc = redis.NewClient(&redis.Options{
			Addr: util.GetRedisUrl(),
			Password: "",
			DB:		  0,
		})
	}

	return Weblensdb{
		mongo: mongodb,
		redis: redisc,
		accessor: username,
	}
}

func (db Weblensdb) GetAccessor() string {
	return db.accessor
}

func (db Weblensdb) GetMedia(fileHash string, includeThumbnail bool) (Media) {
	conn := db.redis.Get(fileHash)
	val, err := conn.Result()
	if err != nil {

		var opts *options.FindOneOptions

		if !includeThumbnail {
			opts = options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: util.IntFromBool(includeThumbnail)}})
		}

		filter := bson.D{{Key: "_id", Value: fileHash}}
		findRet := db.mongo.Collection("images").FindOne(mongo_ctx, filter, opts)

		var i Media
		findRet.Decode(&i)
		return i

	} else {
		//fmt.Println("Redis cache hit")
		var i Media
		json.Unmarshal([]byte(val), &i)
		return i
	}
}

func (db Weblensdb) GetMediaByFilepath(path string, includeThumbnail bool) (Media, error) {
	path, _ = GuaranteeUserRelativePath(path, db.accessor)

	filter := bson.D{{Key: "filepath", Value: path}}

	var opts *options.FindOneOptions
	if !includeThumbnail {
		opts = options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: util.IntFromBool(includeThumbnail)}})
	}
	findRet := db.mongo.Collection("images").FindOne(mongo_ctx, filter, opts)

	if findRet.Err() != nil {
		// util.DisplayError(findRet.Err(), "Failed to get media by filepath")
		return Media{}, fmt.Errorf("failed to get media by filepath (%s): %s", path, findRet.Err())
	}
	var i Media
	findRet.Decode(&i)
	return i, nil
}

// Returns ids of all media in directory with depth 1
func (db Weblensdb) GetMediaInDirectory(dirpath string, recursive bool) ([]Media) {
	var re string
	util.Debug.Println("DIR: ", dirpath)
	relPath, _ := GuaranteeUserRelativePath(dirpath, db.accessor)
	if !recursive {
		re = fmt.Sprintf("^%s\\/?[^\\/]+$", relPath)
	} else {
		re = fmt.Sprintf("^%s/?.*$", relPath)
	}

	// util.Debug.Println("RE", re)

	filter := bson.M{"filepath": bson.M{"$regex": re, "$options": "i"}, "owner": db.accessor}
	opts := options.Find().SetProjection(bson.D{{Key: "thumbnail", Value: 0}})

	findRet, err := db.mongo.Collection("images").Find(mongo_ctx, filter, opts)

	if err != nil {
		panic(err)
	}

	var i []Media
	findRet.All(mongo_ctx, &i)
	return i
}

func (db Weblensdb) GetPagedMedia(sort, owner string, skip, limit int, raw, thumbnails bool) ([]Media, bool) {
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

	userMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "owner", Value: owner}}}}
	pipeline = append(pipeline, userMatchStage)

	skipStage := bson.D{{Key: "$skip", Value: skip}}
	pipeline = append(pipeline, skipStage)

	if limit != 0 {
		limitStage := bson.D{{Key: "$limit", Value: limit}}
		pipeline = append(pipeline, limitStage)
	}

	opts := options.Aggregate()
	cursor, err := db.mongo.Collection("images").Aggregate(mongo_ctx, pipeline, opts)
	if err != nil {
		panic(err)
	}

	var res []Media
	err = cursor.All(mongo_ctx, &res)
	if err != nil {
		panic(err)
	}

	if redisc != nil {
		for i, val := range res {
			go db.redisCacheThumbBytes(val)
			res[i].Thumbnail64 = ""
		}
	}

	return res, len(res) == limit

}

func (db Weblensdb) RedisCacheSet(key string, data string) (error) {
	if redisc == nil {
		return errors.New("redis not initialized")
	}
	_, err := db.redis.Set(key, data, time.Duration(600000000000)).Result()
	return err
}

func (db Weblensdb) RedisCacheGet(key string) (string, error) {
	if redisc == nil {
		return "", errors.New("redis not initialized")
	}
	data, err := db.redis.Get(key).Result()

	return data, err
}

func (db Weblensdb) redisCacheThumbBytes(media Media) (error) {
	if redisc == nil {
		return errors.New("redis not initialized")
	}
	_, err := db.redis.Set(media.FileHash, media, time.Duration(60000000000)).Result()
	return err
}

func (db Weblensdb) DbAddMedia(m *Media) {
	filled, reason := m.IsFilledOut(false)
	if !filled {
		util.Error.Panicf("Refusing to write incomplete media to database for file %s (missing %s)", m.Filepath, reason)
	}

	if (m.Owner == "") {
		owner := GetOwnerFromFilepath(m.Filepath)
		_, err := db.GetUser(owner)
		util.FailOnError(err, "Failed attempt to add media to database (with non existant user)")

		m.Owner = owner
	}

	newPath, err := GuaranteeUserRelativePath(m.Filepath, m.Owner)
	util.FailOnError(err, "Failed to get user relative path when adding media to db")

	m.Filepath = newPath

	// m.Filepath = filepath.Join("/", username, GuaranteeRelativePath(m.Filepath, username))

	filter := bson.D{{Key: "_id", Value: m.FileHash}}
	set := bson.D{{Key: "$set", Value: *m}}
	_, err = db.mongo.Collection("images").UpdateOne(mongo_ctx, filter, set, options.Update().SetUpsert(true))
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) UpdateMediaByFilepath(filepath string, m Media) {
	filepath, _ = GuaranteeUserRelativePath(filepath, db.accessor)
	// m.Filepath = GuaranteeRelativePath(m.Filepath, db.accessor)

	filter := bson.D{{Key: "filepath", Value: filepath}}
	set := bson.D{{Key: "$set", Value: m}}
	_, err := db.mongo.Collection("images").UpdateOne(mongo_ctx, filter, set)
	util.FailOnError(err, "Failed to update media by filepath")
}

func (db Weblensdb) UpdateMediasByFilehash(filehashes []string, newOwner string) {
	user, err := db.GetUser(newOwner)
	util.FailOnError(err, "Failed to get user to update media owner")

	filter := bson.M{"_id": bson.M{"$in": filehashes}}
	update := bson.M{"$set": bson.M{"owner": user.Id}}

	_, err = db.mongo.Collection("images").UpdateMany(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to update media by filehash")
}

func (db Weblensdb) MoveMedia(newFilepath, mediaOwner string, m Media) {
	newFilepath, _ = GuaranteeUserRelativePath(newFilepath, mediaOwner)
	oldFilepath, _ := GuaranteeUserRelativePath(m.Filepath, mediaOwner)

	m.Filepath = newFilepath
	m.GenerateFileHash()

	db.RemoveMediaByFilepath(oldFilepath)
	db.DbAddMedia(&m)
}

func (db Weblensdb) RemoveMediaByFilepath(filepath string) {
	relativePath := GuaranteeRelativePath(filepath)
	filter := bson.D{{Key: "filepath", Value: relativePath}}
	_, err := db.mongo.Collection("images").DeleteOne(mongo_ctx, filter)
	util.FailOnError(err, "Failed to remove media by filepath")
}

func (db Weblensdb) CreateTrashEntry(originalFilepath, trashPath string, mediaData Media) {

	originalFilepath = GuaranteeRelativePath(originalFilepath)

	entry := TrashEntry{OriginalPath: originalFilepath, TrashPath: trashPath, OriginalData: mediaData}

	_, err := db.mongo.Collection("trash").InsertOne(mongo_ctx, entry)
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) GetTrashedFiles() []TrashEntry {
	filter := bson.D{{}}

	ret, err := db.mongo.Collection("trash").Find(mongo_ctx, filter)
	if err != nil {
		panic(err)
	}

	var trashed []TrashEntry
	ret.All(mongo_ctx, &trashed)

	return trashed

}

func (db Weblensdb) RemoveTrashEntry(trashEntry TrashEntry) {
	_, err := db.mongo.Collection("trash").DeleteOne(mongo_ctx, trashEntry)
	if err != nil {
		panic(err)
	}

}

func (db Weblensdb) CheckLogin(username string, password string) bool {
	filter := bson.D{{Key: "username", Value: username}}
	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user User
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
	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to add token to user")
}

func (db Weblensdb) CreateUser(username, password string, admin bool) {
	var user User
	user.Username = username
	user.Admin = admin

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	passHash := string(passHashBytes)
	user.Password = passHash

	user.Id = primitive.ObjectID([]byte(uuid.New().String()))
	_, err = db.mongo.Collection("users").InsertOne(mongo_ctx, user)
	util.FailOnError(err, "Could not add new user")

}

func (db Weblensdb) GetUser(username string) (User, error) {
	filter := bson.D{{Key: "username", Value: username}}

	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user User
	err := ret.Decode(&user)
	if err != nil {
		return user, err
	}
	// util.FailOnError(err, "Could not get user")

	return user, nil
}

func (db Weblensdb) GetUsers() []User {
	filter := bson.D{{}}
	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 0}, {Key: "tokens", Value: 0}, {Key: "password", Value: 0}})

	ret, err := db.mongo.Collection("users").Find(mongo_ctx, filter, opts)
	util.FailOnError(err, "Could not get all users")

	var users []User
	err = ret.All(mongo_ctx, &users)
	util.FailOnError(err, "Could not get users")

	return users
}

func (db Weblensdb) CheckToken(username, token string) bool {
	redisKey := "AuthToken-" + username
	redisRet, _ := db.RedisCacheGet(redisKey)
	if redisRet == token {
		return true
	}

	filter := bson.D{{Key: "username", Value: username}, {Key: "tokens", Value: token}}
	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user User
	err := ret.Decode(&user)
	if err == nil {
		err = db.RedisCacheSet(redisKey, token)
		util.FailOnError(err, "Failed to add webtoken to redis cache")
	}

	return err == nil
}