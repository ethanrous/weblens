package dataStore

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore/album"
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoCtx = context.TODO()

var mongoc *mongo.Client
var mongodb *mongo.Database

func NewDB() *WeblensDB {
	if mongoc == nil {
		var uri = util.GetMongoURI()

		clientOptions := options.Client().ApplyURI(uri).SetTimeout(time.Second)
		var err error
		mongoc, err = mongo.Connect(mongoCtx, clientOptions)
		if err != nil {
			panic(err)
		}
		mongodb = mongoc.Database(util.GetMongoDBName())
		verifyIndexes(mongodb)
	}

	return &WeblensDB{
		mongo: mongodb,
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
		Keys:    bson.M{"contentId": 1},
		Options: &options.IndexOptions{
			// Unique: boolPointer(true),
		},
	}

	_, err = mdb.Collection("media").Indexes().CreateOne(context.TODO(), i)
	if err != nil {
		panic(err)
	}
}

func (db WeblensDB) setMediaHidden(ms []types.Media, hidden bool) error {
	filter := bson.M{"contentId": bson.M{"$in": util.Map(ms, func(m types.Media) types.ContentId { return m.ID() })}}
	update := bson.M{"$set": bson.M{"hidden": hidden}}
	_, err := db.mongo.Collection("media").UpdateMany(mongoCtx, filter, update)
	return err
}

func (db WeblensDB) GetFilteredMedia(
	sort string, requester types.Username, sortDirection int, albumIds []types.AlbumId, raw bool,
) (res []types.Media, err error) {

	var ret *mongo.Cursor

	if len(albumIds) != 0 {
		filter := bson.M{"_id": bson.M{"$in": albumIds}}
		ret, err = db.mongo.Collection("albums").Find(mongoCtx, filter, nil)
	} else {
		filter := bson.M{"$or": bson.A{bson.M{"owner": requester}, bson.M{"sharedWith": requester}}}
		ret, err = db.mongo.Collection("albums").Find(mongoCtx, filter, nil)
	}
	if err != nil {
		return
	}

	var matchedAlbums []album.Album
	err = ret.All(mongoCtx, &matchedAlbums)
	if err != nil || len(matchedAlbums) == 0 {
		return
	}

	var medias []types.Media
	util.Each(
		matchedAlbums, func(a album.Album) {
			medias = append(medias, a.Medias...)
		},
	)

	if len(medias) == 0 {
		return
	}

	medias = util.Filter(
		medias, func(m types.Media) bool {
			if m == nil {
				return false
			} else if !raw {
				return !m.GetMediaType().IsRaw()
			} else {
				return true
			}
		},
	)

	if sort == "createDate" {
		slices.SortFunc(
			medias, func(a, b types.Media) int { return a.GetCreateDate().Compare(b.GetCreateDate()) * sortDirection },
		)
	}
	return
}

func (db WeblensDB) AddMedia(m types.Media) error {
	filled, reason := m.IsFilledOut()
	if !filled {
		err := fmt.Errorf("refusing to write incomplete media to database for media %s (missing %s)", m.ID(), reason)
		return err
	}

	_, err := db.mongo.Collection("media").InsertOne(mongoCtx, m)
	return err
}

func (db WeblensDB) UpdateMedia(m types.Media) error {
	filled, reason := m.IsFilledOut()
	if !filled {
		err := fmt.Errorf("refusing to update incomplete media to database for media %s (missing %s)", m.ID(), reason)
		return err
	}

	filter := bson.M{"contentId": m.ID()}
	update := bson.M{"$set": m}
	_, err := db.mongo.Collection("media").UpdateOne(mongoCtx, filter, update)
	return err
}

func (db WeblensDB) adjustMediaDate(m types.Media, newDate time.Time) error {
	filter := bson.M{"contentId": m.ID()}
	update := bson.M{"$set": bson.M{"createDate": newDate}}
	_, err := db.mongo.Collection("media").UpdateOne(mongoCtx, filter, update)

	return err
}

func (db WeblensDB) deleteMedia(mId types.ContentId) error {
	filter := bson.M{"contentId": mId}
	_, err := db.mongo.Collection("media").DeleteOne(mongoCtx, filter)
	if err != nil {
		return err
	}

	filter = bson.M{"medias": mId}
	update := bson.M{"$pull": bson.M{"medias": mId}}
	_, err = db.mongo.Collection("albums").UpdateMany(mongoCtx, filter, update)
	if err != nil {
		return err
	}

	filter = bson.M{"cover": mId}
	update = bson.M{"$set": bson.M{"cover": ""}}
	_, err = db.mongo.Collection("albums").UpdateMany(mongoCtx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (db WeblensDB) addFileToMedia(m types.Media, f types.WeblensFile) (err error) {
	filter := bson.M{"contentId": m.ID()}
	update := bson.M{"$addToSet": bson.M{"fileIds": f.ID()}}

	_, err = db.mongo.Collection("media").UpdateOne(mongoCtx, filter, update)
	return
}

func (db WeblensDB) removeFileFromMedia(mId types.ContentId, fId types.FileId) (err error) {
	filter := bson.M{"contentId": mId}
	update := bson.M{"$pull": bson.M{"fileIds": fId}}

	_, err = db.mongo.Collection("media").UpdateOne(mongoCtx, filter, update)
	return
}

func (db WeblensDB) newTrashEntry(t trashEntry) error {
	_, err := db.mongo.Collection("trash").InsertOne(mongoCtx, t)
	return err
}

func (db WeblensDB) getTrashEntry(fileId types.FileId) (entry trashEntry, err error) {
	filter := bson.M{"trashFileId": fileId}
	res := db.mongo.Collection("trash").FindOne(mongoCtx, filter)
	if res.Err() != nil {
		err = res.Err()
		return
	}

	err = res.Decode(&entry)
	if err != nil {
		return
	}
	return
}

func (db WeblensDB) removeTrashEntry(trashFileId types.FileId) error {
	filter := bson.M{"trashFileId": trashFileId}
	_, err := db.mongo.Collection("trash").DeleteOne(mongoCtx, filter)

	return err
}

func (db WeblensDB) AddTokenToUser(username types.Username, token string) {
	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "tokens", Value: token}}}}
	_, err := db.mongo.Collection("users").UpdateOne(mongoCtx, filter, update)
	util.FailOnError(err, "Failed to add token to user")
}

func (db WeblensDB) CreateUser(u *user.User) error {
	u.Id = primitive.NewObjectID()
	_, err := db.mongo.Collection("users").InsertOne(mongoCtx, u)
	return err
}

func (db WeblensDB) GetUser(username types.Username) (types.User, error) {
	filter := bson.M{"username": username}

	ret := db.mongo.Collection("users").FindOne(mongoCtx, filter)

	var u user.User
	err := ret.Decode(&u)
	if err != nil {
		return &u, err
	}

	return &u, nil
}

func (db WeblensDB) activateUser(username types.Username) {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"activated": true}}

	_, err := db.mongo.Collection("users").UpdateOne(mongoCtx, filter, update)
	util.FailOnError(err, "Failed to activate user")
}

func (db WeblensDB) deleteUser(username types.Username) {
	filter := bson.M{"username": username}
	_, err := db.mongo.Collection("users").DeleteOne(mongoCtx, filter)
	if err != nil {
		util.ShowErr(err)
	}
}

// func (db WeblensDB) getUsers(ft types.FileTree) ([]user, error) {
// 	filter := bson.D{{}}
//
// 	ret, err := db.mongo.Collection("users").Find(mongoCtx, filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var users []user
// 	err = ret.All(mongoCtx, &users)
//
// 	for _, user := range users {
// 		homePath := filepath.Join(mediaRoot.absolutePath, user.Username.String())
// 		user.HomeFolder = ft.Get(ft.GenerateFileId(homePath))
// 		user.TrashFolder = ft.Get(ft.GenerateFileId(filepath.Join(homePath, ".user_trash")))
// 	}
//
// 	return users, err
// }

func (db WeblensDB) SearchUsers(searchStr string) []types.Username {
	opts := options.Find().SetProjection(bson.M{"username": 1})
	ret, err := db.mongo.Collection("users").Find(
		mongoCtx, bson.M{"username": bson.M{"$regex": searchStr, "$options": "i"}}, opts,
	)
	// ret, err := db.mongo.Collection("users").Find(mongo_ctx, bson.M{})
	if err != nil {
		util.ErrTrace(err, "Failed to autocomplete user search")
		return []types.Username{}
	}

	var users []user.User
	err = ret.All(mongoCtx, &users)
	if err != nil {
		util.ErrTrace(err, "Failed to decode user search")
		return []types.Username{}
	}

	return util.Map(users, func(u user.User) types.Username { return u.Username })
}

// func (db WeblensDB) GetSharedWith(username types.Username) []types.Share {
// 	filter := bson.M{"accessors": username}
// 	ret, err := db.mongo.Collection("shares").Find(mongoCtx, filter)
// 	if err != nil {
// 		util.ErrTrace(err, "Failed to get shared files")
// 		return nil
//
// 	}
//
// 	var fileShares []share.fileShareData
// 	err = ret.All(mongoCtx, &fileShares)
// 	if err != nil {
// 		util.ErrTrace(err, "Failed to get shared files")
// 		return nil
// 	}
//
// 	fileShares = util.Filter(fileShares, func(s share.fileShareData) bool { return s.Enabled })
// 	return util.Map(fileShares, func(s share.fileShareData) types.Share { return &s })
// }

func (db WeblensDB) GetAlbumById(albumId types.AlbumId) (types.Album, error) {
	filter := bson.M{"_id": albumId}
	res := db.mongo.Collection("albums").FindOne(mongoCtx, filter)

	var target *album.Album
	err := res.Decode(target)

	return target, err
}

// func (db WeblensDB) getAllAlbums() (as []AlbumData) {
// 	res, err := db.mongo.Collection("albums").Find(mongoCtx, bson.M{})
// 	if err != nil {
// 		return
// 	}
//
// 	res.All(mongoCtx, &as)
// 	if as == nil {
// 		as = []AlbumData{}
// 	}
// 	return
// }

func (db WeblensDB) GetAlbumsByUser(user types.Username, nameFilter string, includeShared bool) (as []album.Album) {
	var filter bson.M
	if includeShared {
		filter = bson.M{"$or": []bson.M{{"owner": user}, {"sharedWith": user}}, "name": bson.M{"$regex": nameFilter}}
	} else {
		filter = bson.M{"owner": user, "name": bson.M{"$regex": nameFilter}}
	}
	res, err := db.mongo.Collection("albums").Find(mongoCtx, filter)
	if err != nil {
		return
	}

	res.All(mongoCtx, &as)
	if as == nil {
		as = []album.Album{}
	}
	return
}

// func (db WeblensDB) insertAlbum(a AlbumData) error {
// 	_, err := db.mongo.Collection("albums").InsertOne(mongoCtx, a)
// 	return err
// }

// func (db WeblensDB) addMediaToAlbum(albumId types.AlbumId, mediaIds []types.ContentId) (addedCount int, err error) {
// 	if mediaIds == nil {
// 		return addedCount, fmt.Errorf("nil media ids")
// 	}
//
// 	match := bson.M{"_id": albumId}
// 	preFindRes := db.mongo.Collection("albums").FindOne(mongoCtx, match)
// 	if preFindRes.Err() != nil {
// 		return addedCount, preFindRes.Err()
// 	}
// 	var preData AlbumData
// 	preFindRes.Decode(&preData)
//
// 	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mediaIds}}}
// 	res, err := db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
// 	if err != nil {
// 		return
// 	}
// 	if res == nil || res.MatchedCount == 0 {
// 		return addedCount, fmt.Errorf("no matched albums while adding media")
// 	}
//
// 	postFindRes := db.mongo.Collection("albums").FindOne(mongoCtx, match)
// 	if postFindRes.Err() != nil {
// 		return addedCount, postFindRes.Err()
// 	}
// 	var postData AlbumData
// 	postFindRes.Decode(&postData)
//
// 	addedCount = len(postData.Medias) - len(preData.Medias)
//
// 	return
// }

func (db WeblensDB) removeMediaFromAlbum(albumId types.AlbumId, mediaIds []types.ContentId) error {
	if mediaIds == nil {
		return fmt.Errorf("nil media ids")
	}

	match := bson.M{"_id": albumId}
	update := bson.M{"$pull": bson.M{"medias": bson.M{"$in": mediaIds}}}
	res, err := db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
	if err != nil {
		return err
	}

	if res == nil || res.MatchedCount == 0 {
		return fmt.Errorf("no matched albums while removing media")
	}

	return nil
}

// func (db WeblensDB) removeMediaFromAnyAlbum(mediaId types.ContentId) error {
// 	filter := bson.M{"medias": mediaId}
// 	update := bson.M{"$pull": bson.M{"medias": mediaId}}
// 	_, err := db.mongo.Collection("albums").UpdateMany(mongoCtx, filter, update)
// 	return err
// }

// func (db WeblensDB) setAlbumName(albumId types.AlbumId, newName string) (err error) {
// 	match := bson.M{"_id": albumId}
// 	update := bson.M{"$set": bson.M{"name": newName}}
// 	_, err = db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
// 	return
// }

// func (db WeblensDB) SetAlbumCover(albumId types.AlbumId, coverMediaId types.ContentId, prom1, prom2 string) (err error) {
// 	match := bson.M{"_id": albumId}
// 	update := bson.M{"$set": bson.M{"cover": coverMediaId, "primaryColor": prom1, "secondaryColor": prom2}}
// 	_, err = db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
// 	return
// }

// func (db WeblensDB) shareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
// 	match := bson.M{"_id": albumId}
// 	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
// 	_, err = db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
// 	return
// }

// func (db WeblensDB) unshareAlbum(albumId types.AlbumId, users []types.Username) (err error) {
// 	match := bson.M{"_id": albumId}
// 	update := bson.M{"$pull": bson.M{"sharedWith": bson.M{"$in": users}}}
// 	_, err = db.mongo.Collection("albums").UpdateOne(mongoCtx, match, update)
// 	return
// }

// func (db WeblensDB) DeleteAlbum(albumId types.AlbumId) (err error) {
// 	match := bson.M{"_id": albumId}
// 	_, err = db.mongo.Collection("albums").DeleteOne(mongoCtx, match)
// 	return
// }

// func (db WeblensDB) getAllShares() (ss []types.Share, err error) {
// 	ret, err := db.mongo.Collection("shares").Find(mongoCtx, bson.M{"shareType": "file"})
// 	if err != nil {
// 		return
// 	}
// 	var fileShares []*share.fileShareData
// 	ret.All(mongoCtx, &fileShares)
//
// 	ss = append(ss, util.Map(fileShares, func(fs *share.fileShareData) types.Share { return fs })...)
//
// 	return
//
// }

// func (db WeblensDB) removeFileShare(shareId types.ShareId) (err error) {
// 	filter := bson.M{"_id": shareId, "shareType": FileShare}
//
// 	_, err = db.mongo.Collection("shares").DeleteOne(mongoCtx, filter)
//
// 	if errors.Is(err, mongo.ErrNoDocuments) {
// 		err = ErrNoShare
// 		return
// 	}
//
// 	return
// }

// func (db WeblensDB) newFileShare(shareInfo share.fileShareData) (err error) {
//
// 	_, err = db.mongo.Collection("shares").InsertOne(mongoCtx, shareInfo)
//
// 	// This is not good and is not permanent
// 	// Shares will eventually exist within the weblens file so it doesn't
// 	// need to do a db lookup to find if it already exists
// 	// TODO
// 	if mongo.IsDuplicateKeyError(err) {
// 		err = nil
// 	}
//
// 	return
// }

// func (db WeblensDB) updateFileShare(shareId types.ShareId, s *share.fileShareData) (err error) {
// 	filter := bson.M{"_id": shareId, "shareType": "file"}
// 	update := bson.M{"$set": s}
// 	_, err = db.mongo.Collection("shares").UpdateOne(mongoCtx, filter, update)
// 	return
// }

// func (db WeblensDB) getFileShare(shareId types.ShareId) (s share.fileShareData, err error) {
// 	filter := bson.M{"_id": shareId, "shareType": "file"}
// 	ret := db.mongo.Collection("shares").FindOne(mongoCtx, filter)
// 	err = ret.Decode(&s)
// 	return
// }

func (db WeblensDB) newApiKey(keyInfo ApiKeyInfo) error {
	keyInfo.Id = primitive.NewObjectID()
	_, err := db.mongo.Collection("apiKeys").InsertOne(mongoCtx, keyInfo)
	return err
}

func (db WeblensDB) updateUsingKey(key types.WeblensApiKey, serverId string) error {
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"remoteUsing": serverId}}

	_, err := db.mongo.Collection("apiKeys").UpdateOne(mongoCtx, filter, update)
	return err
}

func (db WeblensDB) removeApiKey(key types.WeblensApiKey) {
	filter := bson.M{"key": key}
	db.mongo.Collection("apiKeys").DeleteOne(mongoCtx, filter)
}

func (db WeblensDB) getApiKeysByUser(username types.Username) []ApiKeyInfo {
	filter := bson.M{"owner": username}
	ret, err := db.mongo.Collection("apiKeys").Find(mongoCtx, filter)
	if err != nil {
		util.ErrTrace(err)
		return nil
	}

	var keys []ApiKeyInfo
	ret.All(mongoCtx, &keys)

	return keys
}

func (db WeblensDB) getApiKeys() []ApiKeyInfo {
	ret, err := db.mongo.Collection("apiKeys").Find(mongoCtx, bson.M{})
	if err != nil {
		util.ShowErr(err)
		return nil
	}

	var k []ApiKeyInfo
	err = ret.All(mongoCtx, &k)
	if err != nil {
		util.ShowErr(err)
		return nil
	}

	return k
}

// func (db WeblensDB) newServer(srvI srvInfo) error {
// 	if srvI.IsThisServer {
// 		existing, _ := db.getThisServerInfo()
// 		if existing != nil {
// 			return ErrAlreadyInit
// 		}
// 	}
// 	db.mongo.Collection("servers").InsertOne(mongoCtx, srvI)
// 	return nil
// }

// func (db WeblensDB) getServers() ([]*srvInfo, error) {
// 	ret, err := db.mongo.Collection("servers").Find(mongoCtx, bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	var servers []*srvInfo
// 	err = ret.All(mongoCtx, &servers)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return servers, nil
// }

func (db WeblensDB) removeServer(remoteId string) {
	db.mongo.Collection("servers").DeleteOne(mongoCtx, bson.M{"_id": remoteId})
}

// func (db WeblensDB) getUsingKey(key types.WeblensApiKey) *srvInfo {
// 	filter := bson.M{"usingKey": key}
// 	ret := db.mongo.Collection("servers").FindOne(mongoCtx, filter)
// 	if ret.Err() != nil {
// 		// util.ShowErr(ret.Err())
// 		return nil
// 	}
//
// 	var remote srvInfo
// 	ret.Decode(&remote)
//
// 	return &remote
// }

// func (db WeblensDB) getThisServerInfo() (*srvInfo, error) {
// 	ret := db.mongo.Collection("servers").FindOne(mongoCtx, bson.M{"isThisServer": true})
// 	if ret.Err() != nil {
// 		if errors.Is(ret.Err(), mongo.ErrNoDocuments) {
// 			return nil, types.ErrServerNotInit
// 		}
// 		util.ErrTrace(ret.Err())
// 		return nil, ret.Err()
// 	}
// 	var si srvInfo
// 	ret.Decode(&si)
//
// 	return &si, nil
// }

// func (db WeblensDB) journalSince(since time.Time) ([]types.JournalEntry, error) {
// 	filter := bson.M{"timestamp": bson.M{
// 		"$gt": primitive.NewDateTimeFromTime(since),
// 	}}
// 	opts := options.Find().SetSort(bson.M{"timestamp": -1})
// 	ret, err := db.mongo.Collection("journal").Find(mongoCtx, filter, opts)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var obj []*fileJournalEntry
// 	err = ret.All(mongoCtx, &obj)
//
// 	return util.SliceConvert[types.JournalEntry](obj), err
// }

// func (db WeblensDB) getLatestBackup() (*backupFile, error) {
// 	pipe := bson.A{bson.M{"$sort": bson.M{"events.timestamp": -1}}, bson.M{"$limit": 1}}
// 	ret, err := db.mongo.Collection("backupHistory").Aggregate(mongoCtx, pipe)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !ret.Next(mongoCtx) {
// 		return nil, ErrNoBackup
// 	}
//
// 	obj := backupFile{}
// 	err = ret.Decode(&obj)
// 	return &obj, err
// }

// func (db WeblensDB) newBackupFileRecord(bf *backupFile) error {
// 	_, err := db.mongo.Collection("fileHistory").InsertOne(mongoCtx, bf)
// 	return err
// }

func (db WeblensDB) setContentId(fileId types.FileId, contentId types.ContentId) error {
	filter := bson.M{"fileId": fileId}
	update := bson.M{"$set": bson.M{"contentId": contentId}}
	_, err := db.mongo.Collection("fileHistory").UpdateOne(mongoCtx, filter, update)
	return err
}

// func (db WeblensDB) backupFileAddHist(newFId, oldFId types.FileId, newHistory []types.FileJournalEntry) error {
// 	filter := bson.M{"fileId": oldFId}
// 	update := bson.M{"$push": bson.M{"events": bson.M{"$each": newHistory}}, "$set": bson.M{"fileId": newFId}}
// 	_, err := db.mongo.Collection("fileHistory").UpdateOne(mongoCtx, filter, update)
// 	return err
// }

// func (db WeblensDB) backupRestoreFile(newFId types.FileId, backupId string, newHistory []types.FileJournalEntry) error {
// 	objId, err := primitive.ObjectIDFromHex(backupId)
// 	if err != nil {
// 		return err
// 	}
// 	slices.SortFunc(newHistory, FileJournalEntrySort)
// 	filter := bson.M{"_id": objId}
// 	update := bson.M{"$push": bson.M{"events": bson.M{"$each": newHistory}}, "$set": bson.M{"fileId": newFId, "lastUpdate": newHistory[len(newHistory)-1].JournaledAt()}}
// 	_, err = db.mongo.Collection("fileHistory").UpdateOne(mongoCtx, filter, update)
// 	return err
// }

// func (db WeblensDB) getSnapshots() (jes []types.JournalEntry, err error) {
// 	filter := bson.M{"action": "backup"}
// 	res, err := db.mongo.Collection("journal").Find(mongoCtx, filter)
// 	if err != nil {
// 		return
// 	}
//
// 	var backups []*backupJournalEntry
// 	err = res.All(mongoCtx, &backups)
// 	if err != nil {
// 		return
// 	}
//
// 	jes = util.SliceConvert[types.JournalEntry](backups)
//
// 	return
// }

// func (db WeblensDB) fileEventsByPath(folderPath string) (files []backupFile, err error) {
// 	regex := "^" + folderPath + "[^/]*$"
// 	filter := bson.M{"events.path": bson.M{"$regex": regex}}
// 	ret, err := db.mongo.Collection("fileHistory").Find(mongoCtx, filter)
// 	if err != nil {
// 		return
// 	}
//
// 	err = ret.All(mongoCtx, &files)
// 	return
// }

// func (db WeblensDB) getFilesPathAndTime(folderPath string, before time.Time) (files []backupFile, err error) {
// 	before = before.Truncate(time.Second).Add(time.Second)
//
// 	regex := "^" + folderPath + "[^/]+$"
// 	pipe := bson.A{
// 		bson.M{
// 			"$match": bson.M{
// 				"events": bson.M{
// 					"$elemMatch": bson.M{
// 						"path": bson.M{
// 							"$regex": regex,
// 						},
// 					},
// 					"$not": bson.M{
// 						"$elemMatch": bson.M{
// 							"action": "fileDelete",
// 							"timestamp": bson.M{
// 								"$lt": before,
// 							},
// 						},
// 					},
// 				},
// 				"$expr": bson.M{
// 					"$let": bson.M{
// 						"vars": bson.M{
// 							"lastMove": bson.M{
// 								"$last": bson.M{
// 									"$filter": bson.M{
// 										"input": "$events",
// 										"as":    "event",
// 										"cond": bson.M{
// 											"$and": bson.A{
// 												bson.M{
// 													"$eq": bson.A{
// 														"$$event.action",
// 														"fileMove",
// 													},
// 												},
// 												bson.M{
// 													"$lt": bson.A{
// 														"$$event.timestamp",
// 														before,
// 													},
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 						"in": bson.M{
// 							"$or": bson.A{
// 								bson.M{
// 									"$ne": bson.A{
// 										"$$lastMove.action",
// 										"fileMove",
// 									},
// 								},
// 								bson.M{
// 									"$regexMatch": bson.M{
// 										"input": "$$lastMove.path",
// 										"regex": regex,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		bson.M{
// 			"$addFields": bson.M{
// 				"events": bson.M{
// 					"$filter": bson.M{
// 						"input": "$events",
// 						"as":    "event",
// 						"cond": bson.M{
// 							"$lt": bson.A{
// 								"$$event.timestamp",
// 								before,
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		bson.M{
// 			"$match": bson.M{
// 				"events": bson.M{
// 					"$exists": true,
// 					"$not":    bson.M{"$size": 0},
// 				},
// 			},
// 		},
// 	}
//
// 	ret, err := db.mongo.Collection("fileHistory").Aggregate(mongoCtx, pipe)
// 	if err != nil {
// 		return
// 	}
//
// 	// var thing []any
// 	err = ret.All(mongoCtx, &files)
// 	return
// }

// func (db WeblensDB) findPastFile(fileId types.FileId, timestamp time.Time) (files []backupFile, err error) {
// 	filter := bson.M{"events.fileId": fileId, "events.timestamp": bson.M{"$lte": timestamp}}
// 	ret, err := db.mongo.Collection("fileHistory").Find(mongoCtx, filter)
// 	if err != nil {
// 		return
// 	}
//
// 	err = ret.All(mongoCtx, &files)
// 	return
// }
