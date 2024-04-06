package dataStore

import (
	"errors"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethanrous/bimg"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type Weblensdb struct {
	mongo    *mongo.Database
	useRedis bool
	redis    *redis.Client
}

type weblensFile struct {
	id           types.FileId
	absolutePath string
	filename     string
	owner        types.User
	size         int64
	isDir        *bool
	modifyDate   time.Time
	media        types.Media
	parent       types.WeblensFile

	childLock *sync.Mutex
	children  map[types.FileId]types.WeblensFile

	tasksLock  *sync.Mutex
	tasksUsing []types.Task

	shares []*fileShareData

	// Mark file as read-only internally.
	// This should be checked before any write action is to be taken
	// this should not be changed during run-time, only set in FsInit
	// if a directory is `readOnly`, all children are as well
	readOnly bool
}

type Media struct {
	mediaId          types.MediaId
	fileIds          []types.FileId
	thumbnailCacheId types.FileId
	fullresCacheIds  []types.FileId
	blurHash         string
	owner            types.User
	mediaWidth       int
	mediaHeight      int
	createDate       time.Time
	mimeType         string
	recognitionTags  []string
	pageCount        int

	mediaType *mediaType
	imported  bool
	rotate    string
	imgBytes  []byte
	image     *bimg.Image
	images    []*bimg.Image

	rawExif           map[string]any
	thumbCacheFile    types.WeblensFile
	fullresCacheFiles []types.WeblensFile
}

type marshalableMedia struct {
	MediaId          types.MediaId  `bson:"mediaId" json:"mediaId"`
	FileIds          []types.FileId `bson:"fileIds" json:"fileIds"`
	ThumbnailCacheId types.FileId   `bson:"thumbnailCacheId" json:"thumbnailCacheId"`
	FullresCacheIds  []types.FileId `bson:"fullresCacheIds" json:"fullresCacheIds"`
	BlurHash         string         `bson:"blurHash" json:"blurHash"`
	Owner            types.Username `bson:"owner" json:"owner"`
	MediaWidth       int            `bson:"width" json:"mediaWidth"`
	MediaHeight      int            `bson:"height" json:"mediaHeight"`
	CreateDate       time.Time      `bson:"createDate" json:"createDate"`
	MimeType         string         `bson:"mimeType" json:"mimeType"`
	RecognitionTags  []string       `bson:"recognitionTags" json:"recognitionTags"`
	PageCount        int            `bson:"pageCount" json:"pageCount"` // for pdfs, etc.
}

type mediaType struct {
	mimeType         string
	friendlyName     string
	fileExtension    []string
	isDisplayable    bool
	isRaw            bool
	isVideo          bool
	supportsImgRecog bool
	multiPage        bool
	rawThumbExifKey  string
}

type marshalableMediaType struct {
	MimeType         string
	FriendlyName     string
	FileExtension    []string
	IsDisplayable    bool
	IsRaw            bool
	IsVideo          bool
	SupportsImgRecog bool
	MultiPage        bool
	RawThumbExifKey  string
}

const (
	Thumbnail types.Quality = "thumbnail"
	Fullres   types.Quality = "fullres"
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

type folderData struct {
	FolderId       types.FileId     `bson:"_id" json:"folderId"`
	ParentFolderId types.FileId     `bson:"parentFolderId" json:"parentFolderId"`
	RelPath        string           `bson:"relPath" json:"relPath"`
	SharedWith     []types.Username `bson:"sharedWith" json:"sharedWith"`
	Shares         []fileShareData  `bson:"shares"`
}

const (
	FileShare  types.ShareType = "file"
	AlbumShare types.ShareType = "album"
)

type fileShareData struct {
	ShareId   types.ShareId    `bson:"_id" json:"shareId"`
	FileId    types.FileId     `bson:"fileId" json:"fileId"`
	ShareName string           `bson:"shareName"`
	Owner     types.Username   `bson:"owner"`
	Accessors []types.Username `bson:"accessors"`
	Public    bool             `bson:"public"`
	Wormhole  bool             `bson:"wormhole"`
	Enabled   bool             `bson:"enabled"`
	Expires   time.Time        `bson:"expires"`
	ShareType types.ShareType  `bson:"shareType"`
}

type accessMeta struct {
	shares      []types.Share
	user        types.User
	usingShare  types.Share
	requestMode types.RequestMode
}

const (
	FileGet types.RequestMode = "fileGet"

	// Grant access unconditionally. This is for sending
	// out updates where the user has already subscribed
	// elsewhere, and we just need to format the data for them
	WebsocketFileUpdate types.RequestMode = "wsFileUpdate"
	MarshalFile types.RequestMode = "marshalFile"

	FileSubscribeRequest types.RequestMode = "fileSub"

	ApiKeyCreate types.RequestMode = "apiKeyCreate"
	ApiKeyGet    types.RequestMode = "apiKeyGet"
)

type trashEntry struct {
	OrigParent   types.FileId `bson:"originalParentId"`
	OrigFilename string       `bson:"originalFilename"`
	TrashFileId  types.FileId `bson:"trashFileId"`
}

type AlbumData struct {
	Id             types.AlbumId    `bson:"_id"`
	Name           string           `bson:"name"`
	Owner          types.Username   `bson:"owner"`
	Cover          types.MediaId    `bson:"cover"`
	PrimaryColor   string           `bson:"primaryColor"`
	SecondaryColor string           `bson:"secondaryColor"`
	Medias         []types.MediaId  `bson:"medias"`
	SharedWith     []types.Username `bson:"sharedWith"`
	ShowOnTimeline bool             `bson:"showOnTimeline"`
}

type ApiKeyInfo struct {
	Key         types.WeblensApiKey `bson:"key"`
	Owner       types.Username      `bson:"owner"`
	CreatedTime time.Time           `bson:"createdTime"`
}

var tasker types.TaskerAgent
var globalCaster types.BroadcasterAgent
var voidCaster types.BroadcasterAgent

func SetTasker(d types.TaskerAgent) {
	tasker = d
}

func SetCaster(b types.BroadcasterAgent) {
	globalCaster = b
}

func SetVoidCaster(b types.BroadcasterAgent) {
	voidCaster = b
}

// Errors
type WeblensUserError types.WeblensError
type WeblensFileError types.WeblensError
type AlreadyExistsError WeblensFileError

var ErrNotUsingRedis = errors.New("not using redis")

var ErrDirNotAllowed WeblensFileError = errors.New("attempted to perform action on directory that does not accept directories")
var ErrDirectoryRequired WeblensFileError = errors.New("attempted to perform an action that requires a directory, but found regular file")
var ErrDirAlreadyExists AlreadyExistsError = errors.New("directory already exists in destination location")

var ErrFileAlreadyExists AlreadyExistsError = errors.New("file already exists in destination location")
var ErrNoFile WeblensFileError = errors.New("file does not exist")
var ErrIllegalFileMove WeblensFileError = errors.New("tried to perform illegal file move")
var ErrWriteOnReadOnly WeblensFileError = errors.New("tried to write to read-only file")

var ErrNoUser WeblensUserError = errors.New("user does not exist")
var ErrUserAlreadyExists WeblensUserError = errors.New("cannot create two users with the same username")
var ErrUserNotAuthorized WeblensUserError = errors.New("user does not have access the requested resource")
var ErrNoFileAccess WeblensUserError = errors.New("user does not have access to file")
var ErrBadPassword WeblensUserError = errors.New("password provided does not authenticate user")

var ErrBadRequestMode = errors.New("access struct does not have correct request mode set for the given function")

var ErrNoMedia = errors.New("no media found")

var ErrNoShare = errors.New("no share found")
var ErrBadShareType = errors.New("expected share type does not match given share type")

var ErrUnsupportedImgType error = errors.New("image type is not supported by weblens")
var ErrPageOutOfRange = errors.New("page number does not exist on media")
