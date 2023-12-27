package dataStore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
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

func NewDB(username string) *Weblensdb {
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
		redisc.FlushAll()
	}

	return &Weblensdb{
		mongo: mongodb,
		redis: redisc,
		accessor: username,
	}
}

func (db Weblensdb) GetAccessor() string {
	return db.accessor
}

func (db Weblensdb) GetMedia(fileHash string, includeThumbnail bool) (Media) {
	val, err := db.RedisCacheGet(fileHash)
	if err == nil {
		var i Media
		json.Unmarshal([]byte(val), &i)
		if (!includeThumbnail && i.Thumbnail64 != "") {
			i.Thumbnail64 = ""
		}
		if (i.Thumbnail64 != "" || !includeThumbnail) {
			return i
		}
	}

	var opts *options.FindOneOptions

	if !includeThumbnail {
		opts = options.FindOne().SetProjection(bson.D{{Key: "thumbnail", Value: 0}})
	}

	filter := bson.D{{Key: "fileHash", Value: fileHash}}
	findRet := db.mongo.Collection("media").FindOne(mongo_ctx, filter, opts)

	var i Media
	findRet.Decode(&i)
	b, _ := json.Marshal(i)
	db.RedisCacheSet(fileHash, string(b))
	return i
}

func (db Weblensdb) GetMediaByFile(file *WeblensFileDescriptor, includeThumbnail bool) (Media, error) {
	filter := bson.M{"parentFolderId": file.ParentFolderId, "filename": file.Filename}

	opts :=	options.FindOne()
	if !includeThumbnail {
		opts = options.FindOne().SetProjection(bson.M{"thumbnail": 0})
	}

	ret := db.mongo.Collection("media").FindOne(mongo_ctx, filter, opts)
	if ret.Err() != nil {
		return Media{}, fmt.Errorf("failed to get media by filepath (%s): %s", file.Filename, ret.Err())
	}

	var m Media
	ret.Decode(&m)
	return m, nil
}

// Returns ids of all media in directory with depth 1
func (db Weblensdb) GetMediaInDirectory(dirpath string, recursive bool) ([]Media) {
	var re string
	relPath, _ := GuaranteeUserRelativePath(dirpath, db.accessor)
	if !recursive {
		re = fmt.Sprintf("^%s\\/?[^\\/]+$", relPath)
	} else {
		re = fmt.Sprintf("^%s/?.*$", relPath)
	}

	filter := bson.M{"filepath": bson.M{"$regex": re, "$options": "i"}, "owner": db.accessor}
	opts := options.Find().SetProjection(bson.D{{Key: "thumbnail", Value: 0}})

	findRet, err := db.mongo.Collection("media").Find(mongo_ctx, filter, opts)

	if err != nil {
		panic(err)
	}

	var i []Media
	findRet.All(mongo_ctx, &i)
	return i
}

func (db Weblensdb) GetPagedMedia(sort, owner string, skip, limit int, raw, thumbnails bool) ([]Media) {
	pipeline := mongo.Pipeline{}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: -1}}}}
	pipeline = append(pipeline, sortStage)

	displayableMatchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "mediaType.isdisplayable", Value: true}}}}
	pipeline = append(pipeline, displayableMatchStage)

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
		limitStage := bson.D{{Key: "$limit", Value: limit * 1000}}
		pipeline = append(pipeline, limitStage)
	}

	opts := options.Aggregate()
	cursor, err := db.mongo.Collection("media").Aggregate(mongo_ctx, pipeline, opts)
	if err != nil {
		panic(err)
	}

	var res []Media
	err = cursor.All(mongo_ctx, &res)
	if err != nil {
		panic(err)
	}

	if redisc != nil {
		go func (medias []Media) {
			for _, val := range medias {
				b, _ := json.Marshal(val)
				db.RedisCacheSet(val.FileHash, string(b))
			}
		}(res)
	}


	if !thumbnails {
		noThumbs := util.Map(res, func(m Media) Media {m.Thumbnail64 = ""; return m})
		return noThumbs
	}

	return res

}

func (db Weblensdb) RedisCacheSet(key string, data string) (error) {
	if redisc == nil {
		return errors.New("redis not initialized")
	}
	_, err := db.redis.Set(key, data, time.Duration(time.Minute) * 10).Result()
	return err
}

func (db Weblensdb) RedisCacheGet(key string) (string, error) {
	if redisc == nil {
		return "", errors.New("redis not initialized")
	}
	data, err := db.redis.Get(key).Result()

	return data, err
}

func (db Weblensdb) RedisCacheBust(key string) {
	db.redis.Del(key)
}

func (db Weblensdb) DbAddMedia(m *Media) {
	filled, reason := m.IsFilledOut(false)
	if !filled {
		util.Error.Panicf("Refusing to write incomplete media to database for file %s (missing %s)", m.Filename, reason)
	}

	if (m.Owner == "") {
		// owner := GetOwnerFromFilepath(m.Filepath)
		// _, err := db.GetUser(owner)
		util.Error.Println("Attempt to add media to database with empty user")
		// util.FailOnError(err, )

		// m.Owner = owner
		return
	}

	_, err := db.mongo.Collection("media").InsertOne(mongo_ctx, m)
	if err != nil {
		panic(err)
	}
}

func (db Weblensdb) UpdateMediasByFilehash(filehashes []string, newOwner string) {
	user, err := db.GetUser(newOwner)
	util.FailOnError(err, "Failed to get user to update media owner")

	filter := bson.M{"fileHash": bson.M{"$in": filehashes}}
	update := bson.M{"$set": bson.M{"owner": user.Id}}

	_, err = db.mongo.Collection("media").UpdateMany(mongo_ctx, filter, update)
	util.FailOnError(err, "Failed to update media by filehash")
}

// Processes necessary changes in database after moving a media file on the filesystem.
// This must be called AFTER the file is moved, i.e. `destinationFile` must exist on the fs
func (db Weblensdb) HandleMediaMove(existingMediaFile, destinationFile *WeblensFileDescriptor) error {
	filter := bson.M{"parentFolderId": existingMediaFile.ParentFolderId, "filename": existingMediaFile.Filename}
	res := db.mongo.Collection("media").FindOne(mongo_ctx, filter)
	var m Media
	err := res.Decode(&m)
	if err != nil {return err}

	m.ParentFolder = destinationFile.ParentFolderId
	m.Filename = destinationFile.Filename

	m.GenerateFileHash(destinationFile)

	update := bson.M{"$set": bson.M{"parentFolderId": destinationFile.ParentFolderId, "filename": destinationFile.Filename, "fileHash": m.FileHash}}

	_, err = db.mongo.Collection("media").UpdateOne(mongo_ctx, filter, update)
	return err
}

func (db Weblensdb) RemoveMediaByFilepath(folderId, filename string) error {
	filter := bson.M{"parentFolderId": folderId, "filename": filename}
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

type username struct {
	Username string `bson:"username"`
}

func (db Weblensdb) SearchUsers(searchStr string) []string {
	ret, err := db.mongo.Collection("users").Find(mongo_ctx, bson.M{"username": bson.M{"$regex": searchStr}})
	util.DisplayError(err, "Failed to autocomplete user search")
	var users []username
	ret.All(mongo_ctx, &users)

	return util.Map(users, func(u username) string {return u.Username})
}

type shareData struct {
	ShareId primitive.ObjectID `bson:"_id" json:"shareId"`
	Owner string `bson:"owner" json:"owner"`
	ParentFolderId string `bson:"parentFolderId" json:"parentFolderId"`
	Filename string `bson:"filename" json:"filename"`
}

func (db Weblensdb) ShareFiles(files []*WeblensFileDescriptor, users []string) {
	if len(files) == 0 || len(users) == 0 {
		return
	}
	opts := options.Update().SetUpsert(true)
	for _, file := range files {
		filter := bson.M{"parentFolderId": file.ParentFolderId, "filename": file.Filename, "owner": db.accessor}
		update := bson.M{"$addToSet": bson.M{"users": bson.M{"$each": users}}}
		_, err := db.mongo.Collection("shares").UpdateOne(mongo_ctx, filter, update, opts)
		util.DisplayError(err, "Failed to create or update share")
	}
}

func (db Weblensdb) GetSharedWith(username string) []*WeblensFileDescriptor {
	opts := options.Find().SetProjection(bson.M{"users": 0})
	filter := bson.M{"users": username}
	ret, err := db.mongo.Collection("shares").Find(mongo_ctx, filter, opts)
	util.DisplayError(err, "Failed to get shared files")

	var files []shareData
	ret.All(mongo_ctx, &files)

	return util.Map(files, func(share shareData) *WeblensFileDescriptor {file := GetWFD(share.ParentFolderId, share.Filename); file.Id(); return file})
}

func (db Weblensdb) CanUserAccess(file *WeblensFileDescriptor, username string) bool {
	opts := options.Find().SetProjection(bson.M{"users": 0})
	filter := bson.M{"parentFolderId": file.ParentFolderId, "filename": file.Filename, "users": username}
	ret, _ := db.mongo.Collection("shares").Find(mongo_ctx, filter, opts)
	return ret.RemainingBatchLength() != 0
}

type folderData struct {
	FolderId string `bson:"_id" json:"folderId"`
	ParentId string `bson:"parentId" json:"parentId"`
	DirPath string `bson:"dirPath" json:"dirPath"`
}

func (db Weblensdb) importDirectory(dirPath string) (folderData, error) {
	relPath := GuaranteeRelativePath(dirPath)
	pathHash := util.HashOfString(8, relPath)

	fldr, err := db.getFolderByPath(relPath)
	if err == nil {
		return fldr, nil
	}

	parentPath := filepath.Dir(relPath)
	var parentId string
	if parentPath == "/" {
		parentId = "0"
	} else {
		parent, err := db.getFolderByPath(parentPath)
		if err != nil {
			parent, err = db.importDirectory(parentPath) // Recursively import directories if not found
			if err != nil {
				return folderData{}, err
			}
		}
		parentId = parent.FolderId
	}

	fldr = folderData{FolderId: pathHash, ParentId: parentId, DirPath: relPath}

	_, err = db.mongo.Collection("folders").InsertOne(mongo_ctx, fldr)
	if err != nil {
		util.DisplayError(err, "Error importing directory to database")
		return fldr, err
	}
	return fldr, nil
}

func (db Weblensdb) deleteDirectory(folderId string) error {
	filter := bson.M{"_id": folderId}
	_, err := db.mongo.Collection("folders").DeleteOne(mongo_ctx, filter)
	return err
}

func (db Weblensdb) deleteMediaByFolder(folderId string) error {
	filter := bson.M{"parentFolderId": folderId}
	_, err := db.mongo.Collection("media").DeleteMany(mongo_ctx, filter)
	return err
}

func (db Weblensdb) getFolderById(folderId string) folderData {
	if folderId == "home" {
		util.LazyStackTrace()
		util.Error.Panicf("Db attempt to get folder by `home` id. This should be translated before reaching the database. See trace above")
	}
	filter := bson.M{"_id": folderId}
	ret := db.mongo.Collection("folders").FindOne(mongo_ctx, filter)
	var f folderData
	ret.Decode(&f)

	return f
}

func (db Weblensdb) getFolderByPath(folderPath string) (folderData, error) {
	relPath := GuaranteeRelativePath(folderPath)
	filter := bson.M{"dirPath": relPath}
	ret := db.mongo.Collection("folders").FindOne(mongo_ctx, filter)
	err := ret.Err()
	var f folderData
	if ret.Err() != nil {
		return f, err
	}
	ret.Decode(&f)

	return f, nil
}

func (db Weblensdb) getMediaByPath(parentFolder, filename string) (Media, error) {
	filter := bson.M{"parentFolderId": parentFolder, "filename": filename}
	// opts := options.FindOne().SetProjection(bson.M{"thumbnail": util.IntFromBool(includeThumbnail)})

	// ret := db.mongo.Collection("media").FindOne(mongo_ctx, filter, opts)
	ret := db.mongo.Collection("media").FindOne(mongo_ctx, filter)

	if ret.Err() != nil {
		return Media{}, fmt.Errorf("failed to get media (%s -> %s): %s", parentFolder, filename, ret.Err())
	}
	var i Media
	ret.Decode(&i)
	return i, nil
}

type AlbumData struct {
	Id string `bson:"_id"`
	Name string `bson:"name"`
	Owner string `bson:"owner"`
	Cover string `bson:"cover"`
	PrimaryColor string `bson:"primaryColor"`
	SecondaryColor string `bson:"secondaryColor"`
	Medias []string `bson:"medias"`
	SharedWith []string `bson:"sharedWith"`
	ShowOnTimeline bool `bson:"showOnTimeline"`
}

func (db Weblensdb) GetAlbum(albumId string) (a AlbumData, err error) {
	filter := bson.M{"_id": albumId, "$or": []bson.M{{"owner": db.accessor}, {"sharedWith": db.accessor}}}
	res := db.mongo.Collection("albums").FindOne(mongo_ctx, filter)
	res.Decode(&a)
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
	return
}

func (db Weblensdb) CreateAlbum(name, owner string) {
	a := AlbumData{Id: util.HashOfString(12, fmt.Sprintln(name, owner)), Name: name, Owner: owner, ShowOnTimeline: true, Medias: []string{}}
	db.mongo.Collection("albums").InsertOne(mongo_ctx, a)
}

func (db Weblensdb) AddMediaToAlbum(albumId string, mediaIds []string) error {
	if mediaIds == nil {
		return fmt.Errorf("nil media ids")
	}

	match := bson.M{"_id": albumId}
	update := bson.M{"$addToSet": bson.M{"medias": bson.M{"$each": mediaIds}}}
	res, err := db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	if err != nil {
		return err
	}

	if res == nil || res.MatchedCount == 0 {
		return fmt.Errorf("no matched albums while adding media")
	}

	return nil
}

func (db Weblensdb) SetAlbumCover(albumId, coverMediaId, prom1, prom2 string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$set": bson.M{"cover": coverMediaId, "primaryColor": prom1, "secondaryColor": prom2}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}

func (db Weblensdb) ShareAlbum(albumId string, users []string) (err error) {
	match := bson.M{"_id": albumId}
	update := bson.M{"$addToSet": bson.M{"sharedWith": bson.M{"$each": users}}}
	_, err = db.mongo.Collection("albums").UpdateOne(mongo_ctx, match, update)
	return
}