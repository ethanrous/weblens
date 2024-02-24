package dataStore

import (
	"errors"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/go-redis/redis"
	"github.com/h2non/bimg"
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
	media        *Media
	parent       *WeblensFile

	childLock *sync.Mutex
	children  map[string]*WeblensFile

	tasksLock  *sync.Mutex
	tasksUsing []Task

	shares []shareData
}

type Media struct {
	MediaId          string               `bson:"fileHash" json:"fileHash"`
	FileIds          []string             `bson:"fileIds" json:"fileIds"`
	ThumbnailCacheId string               `bson:"thumbnailCacheId" json:"thumbnailCacheId"`
	FullresCacheId   string               `bson:"fullresCacheId" json:"fullresCacheId"`
	BlurHash         string               `bson:"blurHash" json:"blurHash"`
	Owner            string               `bson:"owner" json:"owner"`
	MediaWidth       int                  `bson:"width" json:"mediaWidth"`
	MediaHeight      int                  `bson:"height" json:"mediaHeight"`
	ThumbWidth       int                  `bson:"thumbWidth" json:"thumbWidth"`
	ThumbHeight      int                  `bson:"thumbHeight" json:"thumbHeight"`
	ThumbLength      int                  `bson:"thumbLength" json:"thumbLength"`
	FullresLength    int                  `bson:"fullresLength" json:"fullresLength"`
	CreateDate       time.Time            `bson:"createDate" json:"createDate"`
	MediaType        *mediaType           `bson:"mediaType" json:"mediaType"`
	SharedWith       []primitive.ObjectID `bson:"sharedWith" json:"sharedWith"`
	RecognitionTags  []string             `bson:"recognitionTags" json:"recognitionTags"`

	imported bool
	rotate   string
	imgBytes []byte
	image    *bimg.Image
	// thumb            image.Image
	rawExif          map[string]any
	thumbCacheFile   *WeblensFile
	fullresCacheFile *WeblensFile
}

type quality string

const (
	Thumbnail quality = "thumbnail"
	Fullres   quality = "fullres"
)

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
	// txt, doc, directory etc.
	Displayable bool `json:"displayable"`

	IsDir            bool        `json:"isDir"`
	Modifiable       bool        `json:"modifiable"`
	Size             int64       `json:"size"`
	ModTime          time.Time   `json:"modTime"`
	Filename         string      `json:"filename"`
	ParentFolderId   string      `json:"parentFolderId"`
	FileFriendlyName string      `json:"fileFriendlyName"`
	Owner            string      `json:"owner"`
	PathFromHome     string      `json:"pathFromHome"`
	MediaData        *Media      `json:"mediaData"`
	Shares           []shareData `json:"shares"`
	Children         []string    `json:"children"`
}

type folderData struct {
	FolderId       string      `bson:"_id" json:"folderId"`
	ParentFolderId string      `bson:"parentFolderId" json:"parentFolderId"`
	RelPath        string      `bson:"relPath" json:"relPath"`
	SharedWith     []string    `bson:"sharedWith" json:"sharedWith"`
	Shares         []shareData `bson:"shares"`
}

type shareData struct {
	ShareId   string `bson:"shareId"`
	ShareName string `bson:"shareName"`
	Public    bool
	Wormhole  bool
	Expires   time.Time
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
	// SetCaster(BroadcasterAgent)

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
	ScanDirectory(directory *WeblensFile, recursive, deep bool, caster BroadcasterAgent) Task

	ScanFile(file *WeblensFile, m *Media, caster BroadcasterAgent) Task
}

func GimmeTask(t Task) {
	util.Debug.Println("I have task")
}

type BroadcasterAgent interface {
	PushFileCreate(newFile *WeblensFile)
	PushFileUpdate(updatedFile *WeblensFile)
	PushFileMove(preMoveFile *WeblensFile, postMoveFile *WeblensFile)
	PushFileDelete(deletedFile *WeblensFile)
	PushTaskUpdate(taskId string, status string, result any)
}

var tasker TaskerAgent
var globalCaster BroadcasterAgent
var voidCaster BroadcasterAgent

func SetTasker(d TaskerAgent) {
	tasker = d
}

func SetCaster(b BroadcasterAgent) {
	globalCaster = b
}

func SetVoidCaster(b BroadcasterAgent) {
	voidCaster = b
}

// Errors
type alreadyExists error

var ErrNotUsingRedis = errors.New("not using redis")
var ErrDirNotAllowed = errors.New("directory not allowed")
var ErrFileAlreadyExists = errors.New("trying create file that already exists")
var ErrNoFile = errors.New("no file found")
var ErrNoMedia = errors.New("no media found")
var ErrNoShare = errors.New("no share found")
var ErrUnsupportedImgType error = errors.New("image type is not supported by weblens")
