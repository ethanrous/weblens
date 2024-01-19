package dataStore

import (
	"sync"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type Weblensdb struct {
	mongo    *mongo.Database
	redis    *redis.Client
	accessor string
}

type WeblensFile struct {
	id           string
	absolutePath string
	filename     string
	owner        string
	guests       []string
	size         int64
	isDir        *bool
	err          error
	media        *Media
	parent       *WeblensFile
	childLock    *sync.Mutex
	children     map[string]*WeblensFile
}

type marshalableWF struct {
	Id             string
	AbsolutePath   string
	Filename       string
	Owner          string
	ParentFolderId string
	Guests         []string
	Size           int64
	IsDir          bool
}

// Structure for safely sending file information to the client
type FileInfo struct {
	Id string `json:"id"`

	// If the media has been loaded into the database, only if it should be.
	// If media is not required to be imported, this will be set true
	Imported bool `json:"imported"`

	// If the content of the file can be displayed visually.
	// Say the file is a jpg, mov, arw, etc. and not a zip,
	// txt, doc etc.
	Displayable bool `json:"displayable"`

	IsDir          bool      `json:"isDir"`
	Modifiable     bool      `json:"modifiable"`
	Size           int64     `json:"size"`
	ModTime        time.Time `json:"modTime"`
	Filename       string    `json:"filename"`
	ParentFolderId string    `json:"parentFolderId"`
	MediaData      *Media    `json:"mediaData"`
	Owner          string    `json:"owner"`
}

type folderData struct {
	FolderId       string   `bson:"_id" json:"folderId"`
	ParentFolderId string   `bson:"parentFolderId" json:"parentFolderId"`
	RelPath        string   `bson:"relPath" json:"relPath"`
	SharedWith     []string `bson:"sharedWith" json:"sharedWith"`
	// Owner string `bson:"owner" json:"owner"`
}

type alreadyExists error

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
	ScanDirectory(directory *WeblensFile, recursive, deep bool)
}

type BroadcasterAgent interface {
	PushFileCreate(newFile *WeblensFile)
	PushFileUpdate(updatedFile *WeblensFile)
	PushFileMove(preMoveFile *WeblensFile, postMoveFile *WeblensFile)
	PushFileDelete(deletedFile *WeblensFile)
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
