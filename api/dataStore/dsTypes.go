package dataStore

import (
	"sync"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type Weblensdb struct {
	mongo    *mongo.Database
	redis    *redis.Client
	accessor string
}

type WeblensFileDescriptor struct {
	id           string
	absolutePath string
	filename     string
	owner        string
	guests       []string
	size         int64
	isDir        *bool
	err          error
	media        *Media
	parent       *WeblensFileDescriptor
	childLock    *sync.Mutex
	children     map[string]*WeblensFileDescriptor
}

type marshalableWFD struct {
	Id             string
	AbsolutePath   string
	Filename       string
	Owner          string
	ParentFolderId string
	Guests         []string
	Size           int64
	IsDir          bool
}

type folderData struct {
	FolderId       string   `bson:"_id" json:"folderId"`
	ParentFolderId string   `bson:"parentFolderId" json:"parentFolderId"`
	RelPath        string   `bson:"relPath" json:"relPath"`
	SharedWith     []string `bson:"sharedWith" json:"sharedWith"`
	// Owner string `bson:"owner" json:"owner"`
}

type AlbumData struct {
	Id             string   `bson:"_id"`
	Name           string   `bson:"name"`
	Owner          string   `bson:"owner"`
	Cover          string   `bson:"cover"`
	PrimaryColor   string   `bson:"primaryColor"`
	SecondaryColor string   `bson:"secondaryColor"`
	Medias         []string `bson:"medias"`
	SharedWith     []string `bson:"sharedWith"`
	ShowOnTimeline bool     `bson:"showOnTimeline"`
}

type TaskerAgent interface {

	// Parameters:
	//
	//	- `directory` : the weblens file descriptor representing the directory to scan
	//
	//	- `recursive` : if true, scan all children of directory recursively
	//
	//	- `deep` : query and sync with the real underlying filesystem for changes not reflected in the current fileTree
	ScanDirectory(directory *WeblensFileDescriptor, recursive, deep bool)
}

type BroadcasterAgent interface {
	PushItemCreate(newFile *WeblensFileDescriptor)
	PushItemUpdate(updatedFile *WeblensFileDescriptor)
	PushItemMove(preMoveFile *WeblensFileDescriptor, postMoveFile *WeblensFileDescriptor)
	PushItemDelete(deletedFile *WeblensFileDescriptor)
}

// Tasker interface for queueing tasks in the task pool
var tasker TaskerAgent
var caster BroadcasterAgent

func SetTasker(d TaskerAgent) {
	tasker = d
}

func SetCaster(b BroadcasterAgent) {
	caster = b
}
