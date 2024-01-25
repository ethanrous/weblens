package dataStore

import (
	"errors"
	"image"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Weblensdb struct {
	mongo    *mongo.Database
	useRedis bool
	redis    *redis.Client
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

	childLock *sync.Mutex
	children  map[string]*WeblensFile

	tasksLock  *sync.Mutex
	tasksUsing []Task
}

type Media struct {
	MediaId     string               `bson:"fileHash" json:"fileHash"`
	FileId      string               `bson:"fileId" json:"fileId"`
	MediaType   *mediaType           `bson:"mediaType" json:"mediaType"`
	BlurHash    string               `bson:"blurHash" json:"blurHash"`
	Thumbnail64 string               `bson:"thumbnail" json:"thumbnail64"`
	MediaWidth  int                  `bson:"width" json:"mediaWidth"`
	MediaHeight int                  `bson:"height" json:"mediaHeight"`
	ThumbWidth  int                  `bson:"thumbWidth" json:"thumbWidth"`
	ThumbHeight int                  `bson:"thumbHeight" json:"thumbHeight"`
	CreateDate  time.Time            `bson:"createDate" json:"createDate"`
	Owner       string               `bson:"owner" json:"owner"`
	SharedWith  []primitive.ObjectID `bson:"sharedWith" json:"sharedWith"`

	image      image.Image
	rawExif    map[string]any
	thumbBytes []byte
	rotate     string
	imported   bool
}

var gexift *exiftool.Exiftool

func SetExiftool(et *exiftool.Exiftool) {
	gexift = et
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

	IsDir            bool      `json:"isDir"`
	Modifiable       bool      `json:"modifiable"`
	Size             int64     `json:"size"`
	ModTime          time.Time `json:"modTime"`
	Filename         string    `json:"filename"`
	ParentFolderId   string    `json:"parentFolderId"`
	MediaData        *Media    `json:"mediaData"`
	FileFriendlyName string    `json:"fileFriendlyName"`
	Owner            string    `json:"owner"`
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

type Task interface {
	TaskId() string
	TaskType() string
	Status() (bool, string)
	GetResult(string) string
	Wait()
	Cancel()

	ReadError() any
}

// Tasker interface for queueing tasks in the task pool
type TaskerAgent interface {

	// Parameters:
	//
	//	- `directory` : the weblens file descriptor representing the directory to scan
	//
	//	- `recursive` : if true, scan all children of directory recursively
	//
	//	- `deep` : query and sync with the real underlying filesystem for changes not reflected in the current fileTree
	ScanDirectory(directory *WeblensFile, recursive, deep bool) Task
}

type BroadcasterAgent interface {
	PushFileCreate(newFile *WeblensFile)
	PushFileUpdate(updatedFile *WeblensFile)
	PushFileMove(preMoveFile *WeblensFile, postMoveFile *WeblensFile)
	PushFileDelete(deletedFile *WeblensFile)
}

var tasker TaskerAgent
var caster BroadcasterAgent

func SetTasker(d TaskerAgent) {
	tasker = d
}

func SetCaster(b BroadcasterAgent) {
	caster = b
}

// Errors
type alreadyExists error

var ErrNotUsingRedis = errors.New("not using redis")
