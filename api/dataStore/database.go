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

func NewDB() *WeblensDB {
	if mongoc == nil {
		var uri = util.GetMongoURI()

		clientOptions := options.Client().ApplyURI(uri).SetTimeout(time.Second)
		var err error
		mongoc, err = mongo.Connect(mongo_ctx, clientOptions)
		if err != nil {
			panic(err)
		}
		mongodb = mongoc.Database(util.GetMongoDBName())
		verifyIndexes(mongodb)
	}
	if redisc == nil && util.ShouldUseRedis() {
		redisc = redis.NewClient(&redis.Options{
			Addr:     util.GetRedisUrl(),
			Password: "",
			DB:       0,
		})
		redisc.FlushAll()
	}

	return &WeblensDB{
		mongo:    mongodb,
		redis:    redisc,
		useRedis: util.ShouldUseRedis(),
	}
}

func verifyIndexes(mdb *mongo.Database) {
	i := mongo.IndexModel{
		Keys: bson.M{"contentId": 1},
		Options: &options.IndexOptions{
			Sparse: boolPointer(true),
		},
	}

	// indexes := mdb.Collection("fileHistory").SearchIndexes()
	// util.Debug.Println(indexes)

	_, err := mdb.Collection("fileHistory").Indexes().CreateOne(context.TODO(), i)
	if err != nil {
		panic(err)
	}

	i = mongo.IndexModel{
		Keys:    bson.M{"mediaId": 1},
		Options: &options.IndexOptions{
			// Unique: boolPointer(true),
		},
	}

	_, err = mdb.Collection("media").Indexes().CreateOne(context.TODO(), i)
	if err != nil {
		panic(err)
	}
}

func (db WeblensDB) getAllMedia() (ms []*media, err error) {
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

	ms = util.Map(marshMs, func(mm marshalableMedia) *media { m := marshalableToMedia(mm); m.SetImported(true); return m })

	return
}

func (db WeblensDB) setMediaHidden(m types.Media, hidden bool) error {
	filter := bson.M{"mediaId": m.Id()}
	update := bson.M{"$set": bson.M{"hidden": hidden}}
	_, err := db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db WeblensDB) GetFilteredMedia(sort string, requester types.Username, sortDirection int, albumIds []types.AlbumId, raw bool) (res []types.Media, err error) {

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

	var mediaIds []types.ContentId
	util.Each(matchedAlbums, func(a AlbumData) { mediaIds = append(mediaIds, a.Medias...) })

	if len(mediaIds) == 0 {
		return
	}

	res = util.Map(mediaIds, func(mId types.ContentId) types.Media {
		m := MediaMapGet(mId)
		// util.ShowErr(err, fmt.Sprint("Failed to get media ", mId))
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

func (db WeblensDB) RedisCacheSet(key string, data string) error {
	if !db.useRedis {
		return nil
	}

	if db.redis == nil {
		return errors.New("redis not initialized")
	}
	_, err := db.redis.Set(key, data, time.Duration(time.Minute*10)).Result()
	return err
}

func (db WeblensDB) RedisCacheGet(key string) (string, error) {
	if !db.useRedis {
		return "", ErrNotUsingRedis
	}

	if db.redis == nil {
		return "", errors.New("redis not initialized")
	}

	data, err := db.redis.Get(key).Result()

	return data, err
}

func (db WeblensDB) AddMedia(m types.Media) error {
	filled, reason := m.IsFilledOut()
	if !filled {
		err := fmt.Errorf("refusing to write incomplete media to database for media %s (missing %s)", m.Id(), reason)
		return err
	}

	_, err := db.mongo.Collection("media").InsertOne(mongo_ctx, m)
	return err
}

func (db WeblensDB) UpdateMedia(m types.Media) error {
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

func (db WeblensDB) deleteMedia(mId types.ContentId) error {
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

func (db WeblensDB) addFileToMedia(m types.Media, f types.WeblensFile) (err error) {
	filter := bson.M{"mediaId": m.Id()}
	update := bson.M{"$addToSet": bson.M{"fileIds": f.Id()}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db WeblensDB) removeFileFromMedia(mId types.ContentId, fId types.FileId) (err error) {
	filter := bson.M{"mediaId": mId}
	update := bson.M{"$pull": bson.M{"fileIds": fId}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db WeblensDB) newTrashEntry(t trashEntry) error {
	_, err := db.mongo.Collection("trash").InsertOne(mongo_ctx, t)
	return err
}

func (db WeblensDB) getTrashEntry(fileId types.FileId) (entry trashEntry, err error) {
	filter := bson.M{"trashFileId": fileId}
	res := db.mongo.Collection("trash").FindOne(mongo_ctx, filter)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	res.Decode(&entry)
	return
}

func (db WeblensDB) removeTrashEntry(trashFileId types.FileId) error {
	filter := bson.M{"trashFileId": trashFileId}
	_, err := db.mongo.Collection("trash").DeleteOne(mongo_ctx, filter)

	return err
}

func (db WeblensDB) AddTokenToUser(username types.Username, token string) {
	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "tokens", Value: token}}}}
	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to add token to user")
}

func (db WeblensDB) CreateUser(u user) error {
	u.Id = primitive.NewObjectID()
	_, err := db.mongo.Collection("users").InsertOne(mongo_ctx, u)
	return err
}

func (db WeblensDB) GetUser(username types.Username) (user, error) {
	filter := bson.M{"username": username}

	ret := db.mongo.Collection("users").FindOne(mongo_ctx, filter)

	var user user
	err := ret.Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (db WeblensDB) activateUser(username types.Username) {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"activated": true}}

	_, err := db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to activate user")
}

func (db WeblensDB) deleteUser(username types.Username) {
	filter := bson.M{"username": username}
	db.mongo.Collection("users").DeleteOne(mongo_ctx, filter)
}

func (db WeblensDB) getUsers() ([]user, error) {
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

func (db WeblensDB) CheckToken(username, token string) bool {
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
		util.FailOnError(err, "Failed to add WebToken to redis cache")
	}

	return err == nil
}

func (db WeblensDB) updateUser(u *user) (err error) {
	db.redis.Del("AuthToken-" + u.Username.String())
	filter := bson.M{"username": u.Username.String()}
	update := bson.M{"$set": u}
	_, err = db.mongo.Collection("users").UpdateOne(mongo_ctx, filter, update)

	return
}

func (db WeblensDB) FlushRedis() {
	db.redis.FlushAll()
}

func (db WeblensDB) SearchUsers(searchStr string) []types.Username {
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

func (db WeblensDB) GetSharedWith(username types.Username) []types.Share {
	filter := bson.M{"accessors": username}
	ret, err := db.mongo.Collection("shares").Find(mongo_ctx, filter)
	util.ErrTrace(err, "Failed to get shared files")

	fileShares := []fileShareData{}
	ret.All(mongo_ctx, &fileShares)

	fileShares = util.Filter(fileShares, func(s fileShareData) bool { return s.Enabled })

	return util.Map(fileShares, func(s fileShareData) types.Share { return &s })
}

func (db WeblensDB) GetAlbum(albumId types.AlbumId) (a *AlbumData, err error) {
	filter := bson.M{"_id": albumId}
	res := db.mongo.Collection("albums").FindOne(mongo_ctx, filter)

	var album AlbumData
	a = &album
	res.Decode(a)
	err = res.Err()
	return
}

func (db WeblensDB) getAllAlbums() (as []AlbumData) {
	res, err := db.mongo.Collection("albums").Find(mongo_ctx, bson.M{})
	if err != nil {
		return
	}

	res.All(mongo_ctx, &as)
	if as == nil {
		as = []AlbumData{}
	}
	return
}

func (db WeblensDB) GetAlbumsByUser(user types.Username, nameFilter string, includeShared bool) (as []AlbumData) {
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

func (db WeblensDB) insertAlbum(a AlbumData) error {
	_, err := db.mongo.Collection("albums").InsertOne(mongo_ctx, a)
	return err
}

func (db WeblensDB) addMediaToAlbum(albumId types.AlbumId, mediaIds []types.ContentId) (addedCount int, err error) {
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

func (db WeblensDB) removeMediaFromAlbum(albumId types.AlbumId, mediaIds []types.ContentId) error {
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

func (db WeblensDB) setAlbumName(albumId types.AlbumId, newName string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"name": newName}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db WeblensDB) SetAlbumCover(albumId types.AlbumId, coverMediaId types.ContentId, prom1, prom2 string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": coverMediaId, "primaryColor": prom1, "secondaryColor": prom2}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db WeblensDB) shareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db WeblensDB) unshareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$pull": bson.M{"sharedWith": bson.M{"$in": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db WeblensDB) DeleteAlbum(albumId types.AlbumId) (err error) {
	match := bson.M{"_id": albumId}
	_, err = db.mongo.Collection("albums").DeleteOne(mongo_ctx, match)
	return
}

func (db WeblensDB) getAllShares() (ss []types.Share, err error) {
	ret, err := db.mongo.Collection("shares").Find(mongo_ctx, bson.M{"shareType": "file"})
	if err != nil {
		return
	}
	var fileShares []*fileShareData
	ret.All(mongo_ctx, &fileShares)

	ss = append(ss, util.Map(fileShares, func(fs *fileShareData) types.Share { return fs })...)

	return

}

func (db WeblensDB) removeFileShare(shareId types.ShareId) (err error) {
	filter := bson.M{"_id": shareId, "shareType": FileShare}

	_, err = db.mongo.Collection("shares").DeleteOne(mongo_ctx, filter)

	if err == mongo.ErrNoDocuments {
		err = ErrNoShare
		return
	}

	return
}

func (db WeblensDB) newFileShare(shareInfo fileShareData) (err error) {

	_, err = db.mongo.Collection("shares").InsertOne(mongo_ctx, shareInfo)

	// This is not good and is not permanent
	// Shares will eventually exist within the weblens file so it doesn't
	// need to do a db lookup to find if it already exists
	// TODO
	if mongo.IsDuplicateKeyError(err) {
		err = nil
	}

	return
}

func (db WeblensDB) updateFileShare(shareId types.ShareId, s *fileShareData) (err error) {
	filter := bson.M{"_id": shareId, "shareType": "file"}
	update := bson.M{"$set": s}
	_, err = db.mongo.Collection("shares").UpdateOne(mongo_ctx, filter, update)
	return
}

func (db WeblensDB) getFileShare(shareId types.ShareId) (s fileShareData, err error) {
	filter := bson.M{"_id": shareId, "shareType": "file"}
	ret := db.mongo.Collection("shares").FindOne(mongo_ctx, filter)
	err = ret.Decode(&s)
	return
}

func (db WeblensDB) newApiKey(keyInfo ApiKeyInfo) error {
	keyInfo.Id = primitive.NewObjectID()
	_, err := db.mongo.Collection("apiKeys").InsertOne(mongo_ctx, keyInfo)
	return err
}

func (db WeblensDB) updateUsingKey(key types.WeblensApiKey, serverId string) error {
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"remoteUsing": serverId}}

	_, err := db.mongo.Collection("apiKeys").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db WeblensDB) removeApiKey(key types.WeblensApiKey) {
	filter := bson.M{"key": key}
	db.mongo.Collection("apiKeys").DeleteOne(mongo_ctx, filter)
}

func (db WeblensDB) getApiKeysByUser(username types.Username) []ApiKeyInfo {
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

func (db WeblensDB) getApiKeys() []ApiKeyInfo {
	ret, err := db.mongo.Collection("apiKeys").Find(mongo_ctx, bson.M{})
	if err != nil {
		util.ShowErr(err)
		return nil
	}

	var k []ApiKeyInfo
	err = ret.All(mongo_ctx, &k)
	if err != nil {
		util.ShowErr(err)
		return nil
	}

	return k
}

func (db WeblensDB) newServer(srvI srvInfo) error {
	if srvI.IsThisServer {
		existing, _ := db.getThisServerInfo()
		if existing != nil {
			return ErrAlreadyInit
		}
	}
	db.mongo.Collection("servers").InsertOne(mongo_ctx, srvI)
	return nil
}

func (db WeblensDB) getServers() ([]*srvInfo, error) {
	ret, err := db.mongo.Collection("servers").Find(mongo_ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var servers []*srvInfo
	err = ret.All(mongo_ctx, &servers)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (db WeblensDB) removeServer(remoteId string) {
	db.mongo.Collection("servers").DeleteOne(mongo_ctx, bson.M{"_id": remoteId})
}

func (db WeblensDB) getUsingKey(key types.WeblensApiKey) *srvInfo {
	filter := bson.M{"usingKey": key}
	ret := db.mongo.Collection("servers").FindOne(mongo_ctx, filter)
	if ret.Err() != nil {
		// util.ShowErr(ret.Err())
		return nil
	}

	var remote srvInfo
	ret.Decode(&remote)

	return &remote
}

func (db WeblensDB) getThisServerInfo() (*srvInfo, error) {
	ret := db.mongo.Collection("servers").FindOne(mongo_ctx, bson.M{"isThisServer": true})
	if ret.Err() != nil {
		if ret.Err() == mongo.ErrNoDocuments {
			return nil, types.ErrServerNotInit
		}
		util.ErrTrace(ret.Err())
		return nil, ret.Err()
	}
	var si srvInfo
	ret.Decode(&si)

	return &si, nil
}

func (db WeblensDB) journalSince(since time.Time) ([]types.JournalEntry, error) {
	filter := bson.M{"timestamp": bson.M{
		"$gt": primitive.NewDateTimeFromTime(since),
	}}
	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	ret, err := db.mongo.Collection("journal").Find(mongo_ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	obj := []*fileJournalEntry{}
	err = ret.All(mongo_ctx, &obj)

	return util.SliceConvert[types.JournalEntry](obj), err
}

func (db WeblensDB) getJournaledFiles() (bs []backupFile, err error) {
	res, err := db.mongo.Collection("fileHistory").Find(mongo_ctx, bson.M{})
	if err != nil {
		return
	}

	err = res.All(mongo_ctx, &bs)
	return
}

func (db WeblensDB) getLatestBackup() (*backupFile, error) {
	pipe := bson.A{bson.M{"$sort": bson.M{"events.timestamp": -1}}, bson.M{"$limit": 1}}
	ret, err := db.mongo.Collection("backupHistory").Aggregate(mongo_ctx, pipe)
	if err != nil {
		return nil, err
	}

	if !ret.Next(mongo_ctx) {
		return nil, ErrNoBackup
	}

	obj := backupFile{}
	err = ret.Decode(&obj)
	return &obj, err
}

func (db WeblensDB) newBackupFileRecord(bf *backupFile) error {
	_, err := db.mongo.Collection("fileHistory").InsertOne(mongo_ctx, bf)
	return err
}

func (db WeblensDB) setContentId(fileId types.FileId, contentId types.ContentId) error {
	filter := bson.M{"fileId": fileId}
	update := bson.M{"$set": bson.M{"contentId": contentId}}
	_, err := db.mongo.Collection("fileHistory").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db WeblensDB) backupFileAddHist(newFId, oldFId types.FileId, newHistory []types.FileJournalEntry) error {
	filter := bson.M{"fileId": oldFId}
	update := bson.M{"$push": bson.M{"events": bson.M{"$each": newHistory}}, "$set": bson.M{"fileId": newFId}}
	_, err := db.mongo.Collection("fileHistory").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db WeblensDB) backupRestoreFile(newFId types.FileId, backupId string, newHistory []types.FileJournalEntry) error {
	objId, err := primitive.ObjectIDFromHex(backupId)
	if err != nil {
		return err
	}
	slices.SortFunc(newHistory, FileJournalEntrySort)
	filter := bson.M{"_id": objId}
	update := bson.M{"$push": bson.M{"events": bson.M{"$each": newHistory}}, "$set": bson.M{"fileId": newFId, "lastUpdate": newHistory[len(newHistory)-1].JournaledAt()}}
	_, err = db.mongo.Collection("fileHistory").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db WeblensDB) getSnapshots() (jes []types.JournalEntry, err error) {
	filter := bson.M{"action": "backup"}
	res, err := db.mongo.Collection("journal").Find(mongo_ctx, filter)
	if err != nil {
		return
	}

	var bjes []*backupJournalEntry
	err = res.All(mongo_ctx, &bjes)
	if err != nil {
		return
	}

	jes = util.SliceConvert[types.JournalEntry](bjes)

	return
}

func (db WeblensDB) fileEventsByPath(folderPath string) (files []backupFile, err error) {
	regex := "^" + folderPath + "[^/]*$"
	filter := bson.M{"events.path": bson.M{"$regex": regex}}
	ret, err := db.mongo.Collection("fileHistory").Find(mongo_ctx, filter)
	if err != nil {
		return
	}

	err = ret.All(mongo_ctx, &files)
	return
}

func (db WeblensDB) getFilesPathAndTime(folderPath string, before time.Time) (files []backupFile, err error) {
	before = before.Truncate(time.Second).Add(time.Second)

	regex := "^" + folderPath + "[^/]+$"
	pipe := bson.A{
		bson.M{
			"$match": bson.M{
				"events": bson.M{
					"$elemMatch": bson.M{
						"path": bson.M{
							"$regex": regex,
						},
					},
					"$not": bson.M{
						"$elemMatch": bson.M{
							"action": "fileDelete",
							"timestamp": bson.M{
								"$lt": before,
							},
						},
					},
				},
				"$expr": bson.M{
					"$let": bson.M{
						"vars": bson.M{
							"lastMove": bson.M{
								"$last": bson.M{
									"$filter": bson.M{
										"input": "$events",
										"as":    "event",
										"cond": bson.M{
											"$and": bson.A{
												bson.M{
													"$eq": bson.A{
														"$$event.action",
														"fileMove",
													},
												},
												bson.M{
													"$lt": bson.A{
														"$$event.timestamp",
														before,
													},
												},
											},
										},
									},
								},
							},
						},
						"in": bson.M{
							"$or": bson.A{
								bson.M{
									"$ne": bson.A{
										"$$lastMove.action",
										"fileMove",
									},
								},
								bson.M{
									"$regexMatch": bson.M{
										"input": "$$lastMove.path",
										"regex": regex,
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"events": bson.M{
					"$filter": bson.M{
						"input": "$events",
						"as":    "event",
						"cond": bson.M{
							"$lt": bson.A{
								"$$event.timestamp",
								before,
							},
						},
					},
				},
			},
		},
		bson.M{
			"$match": bson.M{
				"events": bson.M{
					"$exists": true,
					"$not":    bson.M{"$size": 0},
				},
			},
		},
	}

	ret, err := db.mongo.Collection("fileHistory").Aggregate(mongo_ctx, pipe)
	if err != nil {
		return
	}

	// var thing []any
	err = ret.All(mongo_ctx, &files)
	return
}

func (db WeblensDB) findPastFile(fileId types.FileId, timestamp time.Time) (files []backupFile, err error) {
	filter := bson.M{"events.fileId": fileId, "events.timestamp": bson.M{"$lte": timestamp}}
	ret, err := db.mongo.Collection("fileHistory").Find(mongo_ctx, filter)
	if err != nil {
		return
	}

	err = ret.All(mongo_ctx, &files)
	return
}
