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

var mongo_ctx = context.TODO()

// var redis_ctx = context.TODO()
var mongoc *mongo.Client
var mongodb *mongo.Database
var redisc *redis.Client

func NewDB() *Weblensdb {
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
			Addr:     util.GetRedisUrl(),
			Password: "",
			DB:       0,
		})
		// redisc.FlushAll()
	}

	return &Weblensdb{
		mongo:    mongodb,
		redis:    redisc,
		useRedis: util.ShouldUseRedis(),
	}
}

func (db Weblensdb) GetMedia(mediaId string) (m *Media) {
	m = MediaMapGet(mediaId)
	if m != nil {
		return m
	}

	filter := bson.D{{Key: "fileHash", Value: mediaId}}
	findRet := db.mongo.Collection("media").FindOne(mongo_ctx, filter)
	findRet.Decode(&m)

	if m != nil {
		mediaMapAdd(m)
	}

	return
}

func (db Weblensdb) getMediaByFile(file *WeblensFile) (m *Media, err error) {
	filter := bson.M{"fileId": file.Id()}

	ret := db.mongo.Collection("media").FindOne(mongo_ctx, filter)
	err = ret.Err()
	if err == mongo.ErrNoDocuments {
		err = ErrNoMedia
	}
	if err != nil {
		return
	}

	m = &Media{}
	err = ret.Decode(m)
	m.imported = true
	return
}

func (db Weblensdb) GetFilteredMedia(sort, requester string, sortDirection int, albums []string, raw, thumbnails bool) (res []Media, err error) {
	pipeline := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "mediaType.isdisplayable", Value: true},
					{Key: "$or",
						Value: bson.A{
							bson.D{{Key: "mediaType.israw", Value: false}},
							bson.D{{Key: "mediaType.israw", Value: raw}},
						},
					},
				},
			},
		},
	}

	pipeline = append(pipeline, bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "albums"},
				{Key: "let",
					Value: bson.D{
						{Key: "fileHash", Value: "$fileHash"},
						{Key: "owner", Value: "$owner"},
					},
				},
				{Key: "pipeline",
					Value: bson.A{
						bson.D{
							{Key: "$match",
								Value: bson.D{
									{Key: "$expr",
										Value: bson.D{
											{Key: "$or",
												Value: bson.A{
													bson.D{
														{Key: "$eq",
															Value: bson.A{
																"$$owner",
																requester,
															},
														},
													},
													bson.D{
														{Key: "$and",
															Value: bson.A{
																bson.D{
																	{Key: "$or",
																		Value: bson.A{
																			bson.D{
																				{Key: "$eq",
																					Value: bson.A{
																						"$owner",
																						requester,
																					},
																				},
																			},
																			bson.D{
																				{Key: "$in",
																					Value: bson.A{
																						requester,
																						"$sharedWith",
																					},
																				},
																			},
																		},
																	},
																},
																bson.D{
																	{Key: "$in",
																		Value: bson.A{
																			"$$fileHash",
																			"$medias",
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						bson.D{
							{Key: "$match",
								Value: bson.D{
									{Key: "$expr",
										Value: bson.D{
											{Key: "$or",
												Value: bson.A{
													bson.D{
														{Key: "$eq",
															Value: bson.A{
																albums,
																bson.A{},
															},
														},
													},
													bson.D{
														{Key: "$and",
															Value: bson.A{
																bson.D{
																	{Key: "$in",
																		Value: bson.A{
																			"$$fileHash",
																			"$medias",
																		},
																	},
																},
																bson.D{
																	{Key: "$in",
																		Value: bson.A{
																			"$_id",
																			albums,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 1}}}},
					},
				},
				{Key: "as", Value: "result"},
			},
		},
	})

	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "result", Value: bson.D{{Key: "$ne", Value: bson.A{}}}}}}})
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: sortDirection}}}})

	cursor, err := db.mongo.Collection("media").Aggregate(mongo_ctx, pipeline)
	if err != nil {
		return
	}

	err = cursor.All(mongo_ctx, &res)
	if err != nil {
		return
	}

	if redisc != nil {
		go func(medias []Media) {
			for _, val := range medias {
				b, _ := json.Marshal(val)
				db.RedisCacheSet(val.MediaId, string(b))
			}
		}(res)
	}

	if !thumbnails {
		noThumbs := util.Map(res, func(m Media) Media { m.Thumbnail64 = ""; return m })
		return noThumbs, err
	}

	return

}

func (db Weblensdb) RedisCacheSet(key string, data string) error {
	if !db.useRedis {
		return nil
	}

	if db.redis == nil {
		return errors.New("redis not initialized")
	}
	_, err := db.redis.Set(key, data, time.Duration(time.Minute*10)).Result()
	return err
}

func (db Weblensdb) RedisCacheGet(key string) (string, error) {
	if !db.useRedis {
		return "", ErrNotUsingRedis
	}

	if db.redis == nil {
		return "", errors.New("redis not initialized")
	}

	data, err := db.redis.Get(key).Result()

	return data, err
}

func (db Weblensdb) RedisCacheBust(key string) {
	if !db.useRedis {
		return
	}
	db.redis.Del(key)
}

func (db Weblensdb) AddMedia(m *Media) error {
	filled, reason := m.IsFilledOut(false)
	if !filled {
		err := fmt.Errorf("refusing to write incomplete media to database for file %s (missing %s)", FsTreeGet(m.FileId).String(), reason)
		return err
	}

	mediaMapAdd(m)

	_, err := db.mongo.Collection("media").InsertOne(mongo_ctx, m)
	return err
}

func (db Weblensdb) UpdateMediasById(mediaIds []string, newOwner string) {
	user, err := db.GetUser(newOwner)
	util.FailOnError(err, "Failed to get user to update media owner")

	filter := bson.M{"fileHash": bson.M{"$in": mediaIds}}
	update := bson.M{"$set": bson.M{"owner": user.Id}}

	_, err = db.mongo.Collection("media").UpdateMany(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to update media by filehash")
}

// Processes necessary changes in database after moving a media file on the filesystem.
// This must be called AFTER the file is moved, i.e. `destinationFile` must exist on the fs
func (db Weblensdb) handleMediaMove(oldFile, newFile *WeblensFile) (err error) {
	filter := bson.M{"fileId": oldFile.Id()}
	update := bson.M{"$set": bson.M{"fileId": newFile.Id()}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db Weblensdb) removeMediaByFile(file *WeblensFile) error {
	filter := bson.M{"fileId": file.Id()}
	_, err := db.mongo.Collection("media").DeleteOne(mongo_ctx, filter)
	return err
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

	if !user.Activated {
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
	user.Tokens = []string{}

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

func (db Weblensdb) ActivateUser(username string) {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"activated": true}}

	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to activate user")
}

func (db Weblensdb) DeleteUser(username string) {
	filter := bson.M{"username": username}

	db.mongo.Collection("users").DeleteOne(mongo_ctx, filter)
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

func (db Weblensdb) ClearCache() {
	db.redis.FlushAll()
}

func (db Weblensdb) SearchUsers(searchStr string) []string {
	ret, err := db.mongo.Collection("users").Find(mongo_ctx, bson.M{"username": bson.M{"$regex": searchStr}})
	util.DisplayError(err, "Failed to autocomplete user search")
	var users []struct {
		Username string `bson:"username"`
	}
	ret.All(mongo_ctx, &users)

	return util.Map(users, func(u struct {
		Username string `bson:"username"`
	}) string {
		return u.Username
	})
}

func (db Weblensdb) ShareFiles(files []*WeblensFile, users []string) error {
	for _, file := range files {
		filter := bson.M{"_id": file.Id()}
		update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
		_, err := db.mongo.Collection("folders").UpdateOne(mongo_ctx, filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db Weblensdb) GetSharedWith(username string) []*WeblensFile {
	opts := options.Find().SetProjection(bson.M{"sharedWith": 0})
	filter := bson.M{"sharedWith": username}
	ret, err := db.mongo.Collection("folders").Find(mongo_ctx, filter, opts)
	util.DisplayError(err, "Failed to get shared files")

	var files []folderData
	ret.All(mongo_ctx, &files)

	return util.Map(files, func(share folderData) *WeblensFile { return FsTreeGet(share.FolderId) })
}

func (db Weblensdb) getFileGuests(file *WeblensFile) []string {
	filter := bson.M{"_id": file.Id()}
	ret := db.mongo.Collection("folders").FindOne(mongo_ctx, filter)

	var fd folderData
	ret.Decode(&fd)

	return fd.SharedWith
}

func (db Weblensdb) getAllFolders() (fs []folderData, err error) {
	ret, err := db.mongo.Collection("folders").Find(mongo_ctx, bson.M{})
	if err != nil {
		util.DisplayError(err)
		return
	}

	err = ret.All(mongo_ctx, &fs)
	return
}

func (db Weblensdb) writeFolder(folder *WeblensFile) error {
	opts := options.Update().SetUpsert(true)

	filter := bson.M{"_id": folder.Id()}
	fldrSet := bson.M{"$set": folderData{FolderId: folder.Id(), ParentFolderId: folder.parent.Id(), RelPath: GuaranteeRelativePath(folder.absolutePath), SharedWith: []string{}}}

	_, err := db.mongo.Collection("folders").UpdateOne(mongo_ctx, filter, fldrSet, opts)
	if err != nil {
		util.DisplayError(err, "Error importing directory to database")
		return err
	}
	return nil
}

func (db Weblensdb) deleteFolder(folder *WeblensFile) error {
	filter := bson.M{"_id": folder.Id()}

	_, err := db.mongo.Collection("folders").DeleteOne(mongo_ctx, filter)
	if err != nil {
		util.DisplayError(err, "Error deleting directory from database")
		return err
	}
	return nil
}

func (db Weblensdb) getFolderById(folderId string) (f folderData) {
	if folderId == "home" {
		util.LazyStackTrace()
		util.Error.Panicf("Db attempt to get folder by `home` id. This should be translated before reaching the database. See trace above")
	}
	filter := bson.M{"_id": folderId}
	ret := db.mongo.Collection("folders").FindOne(mongo_ctx, filter)
	ret.Decode(&f)

	return
}

func (db Weblensdb) GetAlbum(albumId string) (a *AlbumData, err error) {
	filter := bson.M{"_id": albumId}
	res := db.mongo.Collection("albums").FindOne(mongo_ctx, filter)

	var album AlbumData
	a = &album
	res.Decode(a)
	err = res.Err()
	return
}

func (db Weblensdb) GetAlbumsByUser(user, nameFilter string, includeShared bool) (as []AlbumData) {
	var filter bson.M
	if includeShared {
		filter = bson.M{"$or": []bson.M{{"owner": user}, {"sharedWith": user}}, "name": bson.M{"$regex": nameFilter}}
	} else {
		filter = bson.M{"owner": user, "name": bson.M{"$regex": nameFilter}}
	}
	res, err := db.mongo.Collection("albums").Find(mongo_ctx, filter)
	if err != nil {
		return
	}

	res.All(mongo_ctx, &as)
	if as == nil {
		as = []AlbumData{}
	}
	return
}

func (db Weblensdb) CreateAlbum(name, owner string) {
	a := AlbumData{Id: util.HashOfString(12, fmt.Sprintln(name, owner)), Name: name, Owner: owner, ShowOnTimeline: true, Medias: []string{}, SharedWith: []string{}}
	db.mongo.Collection("albums").InsertOne(mongo_ctx, a)
}

func (db Weblensdb) addMediaToAlbum(albumId string, mediaIds []string) (addedCount int, err error) {
	if mediaIds == nil {
		return addedCount, fmt.Errorf("nil media ids")
	}

	match := bson.M{"_id": albumId}
	preFindRes := db.mongo.Collection("albums").FindOne(mongo_ctx, match)
	if preFindRes.Err() != nil {
		return addedCount, preFindRes.Err()
	}
	var preData AlbumData
	preFindRes.Decode(&preData)

	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mediaIds}}}
	res, err := db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	if err != nil {
		return
	}
	if res == nil || res.MatchedCount == 0 {
		return addedCount, fmt.Errorf("no matched albums while adding media")
	}

	postFindRes := db.mongo.Collection("albums").FindOne(mongo_ctx, match)
	if postFindRes.Err() != nil {
		return addedCount, postFindRes.Err()
	}
	var postData AlbumData
	postFindRes.Decode(&postData)

	addedCount = len(postData.Medias) - len(preData.Medias)

	return
}

func (db Weblensdb) removeMediaFromAlbum(albumId string, mediaIds []string) error {
	if mediaIds == nil {
		return fmt.Errorf("nil media ids")
	}

	match := bson.M{"_id": albumId}
	update := bson.M{"$pull": bson.M{"medias": bson.M{"$in": mediaIds}}}
	res, err := db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	if err != nil {
		return err
	}

	if res == nil || res.MatchedCount == 0 {
		return fmt.Errorf("no matched albums while removing media")
	}

	return nil
}

func (db Weblensdb) setAlbumName(albumId, newName string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"name": newName}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) SetAlbumCover(albumId, coverMediaId, prom1, prom2 string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": coverMediaId, "primaryColor": prom1, "secondaryColor": prom2}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) shareAlbum(albumId string, users []string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) unshareAlbum(albumId string, users []string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$pull": bson.M{"sharedWith": bson.M{"$in": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) DeleteAlbum(albumId string) (err error) {
	match := bson.M{"_id": albumId}
	_, err = db.mongo.Collection("albums").DeleteOne(mongo_ctx, match)
	return
}
