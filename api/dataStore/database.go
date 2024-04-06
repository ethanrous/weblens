package dataStore

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		redisc.FlushAll()
	}

	return &Weblensdb{
		mongo:    mongodb,
		redis:    redisc,
		useRedis: util.ShouldUseRedis(),
	}
}

func (db Weblensdb) getAllMedia() (ms []*Media, err error) {
	filter := bson.D{}
	findRet, err := db.mongo.Collection("media").Find(mongo_ctx, filter)
	if err != nil {
		return
	}
	var marshMs []marshalableMedia
	err = findRet.All(mongo_ctx, &marshMs)
	if err != nil {
		return
	}

	ms = util.Map(marshMs, func(mm marshalableMedia) *Media { m := marshalableToMedia(mm); m.SetImported(true); return m })

	return
}

func (db Weblensdb) GetFilteredMedia(sort string, requester types.Username, sortDirection int, albumIds []types.AlbumId, raw bool) (res []types.Media, err error) {

	var ret *mongo.Cursor
	res = []types.Media{}

	if len(albumIds) != 0 {
		filter := bson.M{"_id": bson.M{"$in": albumIds}}
		ret, err = db.mongo.Collection("albums").Find(mongo_ctx, filter, nil)
	} else {
		filter := bson.M{"$or": bson.A{bson.M{"owner": requester}, bson.M{"sharedWith": requester}}}
		ret, err = db.mongo.Collection("albums").Find(mongo_ctx, filter, nil)
	}
	if err != nil {
		return
	}

	var matchedAlbums []AlbumData
	err = ret.All(mongo_ctx, &matchedAlbums)
	if err != nil || len(matchedAlbums) == 0 {
		return
	}

	var mediaIds []types.MediaId
	util.Each(matchedAlbums, func(a AlbumData) { mediaIds = append(mediaIds, a.Medias...) })

	if len(mediaIds) == 0 {
		return
	}

	res = util.Map(mediaIds, func(mId types.MediaId) types.Media {
		m, err := MediaMapGet(mId)
		util.ShowErr(err, fmt.Sprint("Failed to get media ", mId))
		return m
	})

	res = util.Filter(res, func(m types.Media) bool {
		if m == nil {
			return false
		} else if !raw {
			return !m.GetMediaType().IsRaw()
		} else {
			return true
		}
	})

	if sort == "createDate" {
		slices.SortFunc(res, func(a, b types.Media) int { return a.GetCreateDate().Compare(b.GetCreateDate()) * sortDirection })
	}
	return
}

// func (db Weblensdb) _old_GetFilteredMedia(sort, requester string, sortDirection int, albums []string, raw bool) (res []Media, err error) {
// 	pipeline := bson.A{
// 		bson.D{
// 			{Key: "$match",
// 				Value: bson.D{
// 					{Key: "mediaType.isdisplayable", Value: true},
// 					{Key: "$or",
// 						Value: bson.A{
// 							bson.D{{Key: "mediaType.israw", Value: false}},
// 							bson.D{{Key: "mediaType.israw", Value: raw}},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	pipeline = append(pipeline, bson.D{
// 		{Key: "$lookup",
// 			Value: bson.D{
// 				{Key: "from", Value: "albums"},
// 				{Key: "let",
// 					Value: bson.D{
// 						{Key: "mediaId", Value: "$mediaId"},
// 						{Key: "owner", Value: "$owner"},
// 					},
// 				},
// 				{Key: "pipeline",
// 					Value: bson.A{
// 						bson.D{
// 							{Key: "$match",
// 								Value: bson.D{
// 									{Key: "$expr",
// 										Value: bson.D{
// 											{Key: "$or",
// 												Value: bson.A{
// 													bson.D{
// 														{Key: "$eq",
// 															Value: bson.A{
// 																"$$owner",
// 																requester,
// 															},
// 														},
// 													},
// 													bson.D{
// 														{Key: "$and",
// 															Value: bson.A{
// 																bson.D{
// 																	{Key: "$or",
// 																		Value: bson.A{
// 																			bson.D{
// 																				{Key: "$eq",
// 																					Value: bson.A{
// 																						"$owner",
// 																						requester,
// 																					},
// 																				},
// 																			},
// 																			bson.D{
// 																				{Key: "$in",
// 																					Value: bson.A{
// 																						requester,
// 																						"$sharedWith",
// 																					},
// 																				},
// 																			},
// 																		},
// 																	},
// 																},
// 																bson.D{
// 																	{Key: "$in",
// 																		Value: bson.A{
// 																			"$$mediaId",
// 																			"$medias",
// 																		},
// 																	},
// 																},
// 															},
// 														},
// 													},
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{Key: "$match",
// 								Value: bson.D{
// 									{Key: "$expr",
// 										Value: bson.D{
// 											{Key: "$or",
// 												Value: bson.A{
// 													bson.D{
// 														{Key: "$eq",
// 															Value: bson.A{
// 																albums,
// 																bson.A{},
// 															},
// 														},
// 													},
// 													bson.D{
// 														{Key: "$and",
// 															Value: bson.A{
// 																bson.D{
// 																	{Key: "$in",
// 																		Value: bson.A{
// 																			"$$mediaId",
// 																			"$medias",
// 																		},
// 																	},
// 																},
// 																bson.D{
// 																	{Key: "$in",
// 																		Value: bson.A{
// 																			"$_id",
// 																			albums,
// 																		},
// 																	},
// 																},
// 															},
// 														},
// 													},
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 						bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 1}}}},
// 					},
// 				},
// 				{Key: "as", Value: "result"},
// 			},
// 		},
// 	})

// 	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "result", Value: bson.D{{Key: "$ne", Value: bson.A{}}}}}}})
// 	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: sortDirection}}}})

// 	cursor, err := db.mongo.Collection("media").Aggregate(mongo_ctx, pipeline)
// 	if err != nil {
// 		return
// 	}

// 	err = cursor.All(mongo_ctx, &res)
// 	if err != nil {
// 		return
// 	}
// 	if res == nil {
// 		res = []Media{}
// 	}

// 	if redisc != nil {
// 		go func(medias []Media) {
// 			for _, val := range medias {
// 				b, _ := json.Marshal(val)
// 				db.RedisCacheSet(val.MediaId, string(b))
// 			}
// 		}(res)
// 	}

// 	return
// }

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

func (db Weblensdb) AddMedia(m types.Media) error {
	filled, reason := m.IsFilledOut()
	if !filled {
		err := fmt.Errorf("refusing to write incomplete media to database for media %s (missing %s)", m.Id(), reason)
		return err
	}

	_, err := db.mongo.Collection("media").InsertOne(mongo_ctx, m)
	return err
}

func (db Weblensdb) UpdateMedia(m types.Media) error {
	filled, reason := m.IsFilledOut()
	if !filled {
		err := fmt.Errorf("refusing to update incomplete media to database for media %s (missing %s)", m.Id(), reason)
		return err
	}

	filter := bson.M{"mediaId": m.Id()}
	update := bson.M{"$set": m}
	_, err := db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db Weblensdb) deleteMedia(mId types.MediaId) error {
	filter := bson.M{"mediaId": mId}
	_, err := db.mongo.Collection("media").DeleteOne(mongo_ctx, filter)
	if err != nil {
		return err
	}

	filter = bson.M{"medias": mId}
	update := bson.M{"$pull": bson.M{"medias": mId}}
	_, err = db.mongo.Collection("albums").UpdateMany(mongo_ctx, filter, update)
	if err != nil {
		return err
	}

	filter = bson.M{"cover": mId}
	update = bson.M{"$set": bson.M{"cover": ""}}
	_, err = db.mongo.Collection("albums").UpdateMany(mongo_ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// func (db Weblensdb) deleteManyMedias(ms []string) error {
// 	filter := bson.M{"mediaId": bson.M{"$in": ms}}
// 	_, err := db.mongo.Collection("media").DeleteMany(mongo_ctx, filter)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (db Weblensdb) UpdateMediasById(mediaIds []types.MediaId, newOwner types.Username) {
	user, err := db.GetUser(newOwner)
	util.FailOnError(err, "Failed to get user to update media owner")

	filter := bson.M{"mediaId": bson.M{"$in": mediaIds}}
	update := bson.M{"$set": bson.M{"owner": user.Id}}

	_, err = db.mongo.Collection("media").UpdateMany(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to update media by mediaId")
}

func (db Weblensdb) addFileToMedia(m types.Media, f types.WeblensFile) (err error) {
	filter := bson.M{"mediaId": m.Id()}
	update := bson.M{"$addToSet": bson.M{"fileIds": f.Id()}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db Weblensdb) removeFileFromMedia(mId types.MediaId, fId types.FileId) (err error) {
	filter := bson.M{"mediaId": mId}
	update := bson.M{"$pull": bson.M{"fileIds": fId}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return
}

// func (db Weblensdb) getTrashedFiles() []TrashEntry {
// 	filter := bson.D{{}}

// 	ret, err := db.mongo.Collection("trash").Find(mongo_ctx, filter)
// 	if err != nil {
// 		panic(err)
// 	}

// 	var trashed []TrashEntry
// 	ret.All(mongo_ctx, &trashed)

// 	return trashed
// }

func (db Weblensdb) newTrashEntry(t trashEntry) error {
	_, err := db.mongo.Collection("trash").InsertOne(mongo_ctx, t)
	return err
}

func (db Weblensdb) getTrashEntry(fileId types.FileId) (entry trashEntry, err error) {
	filter := bson.M{"trashFileId": fileId}
	res := db.mongo.Collection("trash").FindOne(mongo_ctx, filter)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	res.Decode(&entry)
	return
}

func (db Weblensdb) removeTrashEntry(trashFileId types.FileId) error {
	filter := bson.M{"trashFileId": trashFileId}
	_, err := db.mongo.Collection("trash").DeleteOne(mongo_ctx, filter)

	return err
}

func (db Weblensdb) AddTokenToUser(username types.Username, token string) {
	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "tokens", Value: token}}}}
	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to add token to user")
}

func (db Weblensdb) CreateUser(u user) error {
	u.Id = primitive.NewObjectID()
	_, err := db.mongo.Collection("users").InsertOne(mongo_ctx, u)
	return err
}

func (db Weblensdb) GetUser(username types.Username) (user, error) {
	filter := bson.M{"username": username}

	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user user
	err := ret.Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (db Weblensdb) ActivateUser(username types.Username) {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"activated": true}}

	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to activate user")
}

func (db Weblensdb) deleteUser(username types.Username) {
	filter := bson.M{"username": username}
	db.mongo.Collection("users").DeleteOne(mongo_ctx, filter)
}

func (db Weblensdb) getUsers() ([]user, error) {
	filter := bson.D{{}}
	// opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 0}, {Key: "tokens", Value: 0}, {Key: "password", Value: 0}})

	ret, err := db.mongo.Collection("users").Find(mongo_ctx, filter)
	if err != nil {
		return nil, err
	}

	var users []user
	err = ret.All(mongo_ctx, &users)

	return users, err
}

func (db Weblensdb) CheckToken(username, token string) bool {
	redisKey := "AuthToken-" + username
	redisRet, _ := db.RedisCacheGet(redisKey)
	if redisRet == token {
		return true
	}

	filter := bson.D{{Key: "username", Value: username}, {Key: "tokens", Value: token}}
	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user user
	err := ret.Decode(&user)
	if err == nil {
		err = db.RedisCacheSet(redisKey, token)
		util.FailOnError(err, "Failed to add webtoken to redis cache")
	}

	return err == nil
}

func (db Weblensdb) updateUser(u *user) (err error) {
	db.redis.Del("AuthToken-" + u.Username.String())
	filter := bson.M{"username": u.Username.String()}
	update := bson.M{"$set": u}
	_, err = db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)

	return
}

func (db Weblensdb) FlushRedis() {
	db.redis.FlushAll()
}

func (db Weblensdb) SearchUsers(searchStr string) []types.Username {
	opts := options.Find().SetProjection(bson.M{"username": 1})
	ret, err := db.mongo.Collection("users").Find(mongo_ctx, bson.M{"username": bson.M{"$regex": searchStr, "$options": "i"}}, opts)
	// ret, err := db.mongo.Collection("users").Find(mongo_ctx, bson.M{})
	if err != nil {
		util.ErrTrace(err, "Failed to autocomplete user search")
		return []types.Username{}
	}

	users := []user{}
	ret.All(mongo_ctx, &users)

	return util.Map(users, func(u user) types.Username { return u.Username })
}

func (db Weblensdb) GetSharedWith(username types.Username) []types.Share {
	filter := bson.M{"accessors": username}
	ret, err := db.mongo.Collection("shares").Find(mongo_ctx, filter)
	util.ErrTrace(err, "Failed to get shared files")

	fileShares := []fileShareData{}
	ret.All(mongo_ctx, &fileShares)

	fileShares = util.Filter(fileShares, func(s fileShareData) bool { return s.Enabled })

	return util.Map(fileShares, func(s fileShareData) types.Share { return &s })
}

func (db Weblensdb) GetAlbum(albumId types.AlbumId) (a *AlbumData, err error) {
	filter := bson.M{"_id": albumId}
	res := db.mongo.Collection("albums").FindOne(mongo_ctx, filter)

	var album AlbumData
	a = &album
	res.Decode(a)
	err = res.Err()
	return
}

func (db Weblensdb) GetAlbumsByUser(user types.Username, nameFilter string, includeShared bool) (as []AlbumData) {
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

func (db Weblensdb) CreateAlbum(name string, owner types.Username) {
	a := AlbumData{Id: types.AlbumId(util.GlobbyHash(12, fmt.Sprintln(name, owner))), Name: name, Owner: owner, ShowOnTimeline: true, Medias: []types.MediaId{}, SharedWith: []types.Username{}}
	db.mongo.Collection("albums").InsertOne(mongo_ctx, a)
}

func (db Weblensdb) addMediaToAlbum(albumId types.AlbumId, mediaIds []types.MediaId) (addedCount int, err error) {
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

func (db Weblensdb) removeMediaFromAlbum(albumId types.AlbumId, mediaIds []types.MediaId) error {
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

func (db Weblensdb) setAlbumName(albumId types.AlbumId, newName string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"name": newName}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) SetAlbumCover(albumId types.AlbumId, coverMediaId types.MediaId, prom1, prom2 string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": coverMediaId, "primaryColor": prom1, "secondaryColor": prom2}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) shareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) unshareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$pull": bson.M{"sharedWith": bson.M{"$in": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) DeleteAlbum(albumId types.AlbumId) (err error) {
	match := bson.M{"_id": albumId}
	_, err = db.mongo.Collection("albums").DeleteOne(mongo_ctx, match)
	return
}

func (db Weblensdb) getAllShares() (ss []types.Share, err error) {
	ret, err := db.mongo.Collection("shares").Find(mongo_ctx, bson.M{"shareType": "file"})
	if err != nil {
		return
	}
	var fileShares []*fileShareData
	ret.All(mongo_ctx, &fileShares)

	ss = append(ss, util.Map(fileShares, func(fs *fileShareData) types.Share { return fs })...)

	return

}

func (db Weblensdb) removeFileShare(shareId types.ShareId) (err error) {
	filter := bson.M{"_id": shareId, "shareType": FileShare}

	_, err = db.mongo.Collection("shares").DeleteOne(mongo_ctx, filter)

	if err == mongo.ErrNoDocuments {
		err = ErrNoShare
		return
	}

	return
}

func (db Weblensdb) newFileShare(shareInfo fileShareData) (err error) {

	_, err = db.mongo.Collection("shares").InsertOne(mongo_ctx, shareInfo)

	// This is not good and is not permenant
	// Shares will eventually exist within the weblens file so it doesn't
	// need to do a db lookup to find if it already exists
	// TODO
	if mongo.IsDuplicateKeyError(err) {
		err = nil
	}

	return
}

func (db Weblensdb) updateFileShare(shareId types.ShareId, s *fileShareData) (err error) {
	filter := bson.M{"_id": shareId, "shareType": "file"}
	update := bson.M{"$set": s}
	_, err = db.mongo.Collection("shares").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db Weblensdb) getFileShare(shareId types.ShareId) (s fileShareData, err error) {
	filter := bson.M{"_id": shareId, "shareType": "file"}
	ret := db.mongo.Collection("shares").FindOne(mongo_ctx, filter)
	err = ret.Decode(&s)
	return
}

func (db Weblensdb) newApiKey(key ApiKeyInfo) {
	db.mongo.Collection("apiKeys").InsertOne(mongo_ctx, key)
}

func (db Weblensdb) getApiKeysByUser(username types.Username) []ApiKeyInfo {
	filter := bson.M{"owner": username}
	ret, err := db.mongo.Collection("apiKeys").Find(mongo_ctx, filter)
	if err != nil {
		util.ErrTrace(err)
		return nil
	}

	var keys []ApiKeyInfo
	ret.All(mongo_ctx, &keys)

	return keys
}

func (db Weblensdb) getApiKey(key string) ApiKeyInfo {
	filter := bson.M{"key": key}
	ret := db.mongo.Collection("apiKeys").FindOne(mongo_ctx, filter)
	if ret.Err() != nil {
		util.ErrTrace(ret.Err())
		return ApiKeyInfo{}
	}

	var k ApiKeyInfo
	ret.Decode(&k)

	return k
}