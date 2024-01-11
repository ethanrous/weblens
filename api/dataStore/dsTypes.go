package dataStore

import (
	"sync"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type Weblensdb struct {
	mongo 		*mongo.Database
	redis 		*redis.Client
	accessor 	string
}

type WeblensFileDescriptor struct {
	id string
	absolutePath string
	filename string
	owner string
	ParentFolderId string
	guests []string
	size int64

	isDir *bool
	err error

	media *Media

	parent *WeblensFileDescriptor
	childLock *sync.Mutex
	children map[string]*WeblensFileDescriptor
}

type marshalableWFD struct {
	Id string
	AbsolutePath string
	Filename string
	Owner string
	ParentFolderId string
	Guests []string
	Size int64
	IsDir bool
}

type folderData struct {
	FolderId string `bson:"_id" json:"folderId"`
	// Owner string `bson:"owner" json:"owner"`
	ParentFolderId string `bson:"parentFolderId" json:"parentFolderId"`
	RelPath string `bson:"relPath" json:"relPath"`
	SharedWith []string `bson:"sharedWith" json:"sharedWith"`
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