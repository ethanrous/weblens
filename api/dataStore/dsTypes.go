package dataStore

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// type initStore struct {
// }

type coreStore struct {
}

type backupStore struct {
	req types.Requester
}

func NewStore(req types.Requester) types.Store {
	if thisServer == nil || thisServer.Role == types.Core || thisServer.Role == types.Initialization {
		return &coreStore{}
	} else {
		return &backupStore{
			req: req,
		}
	}
}

type srvInfo struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`

	// apiKey that remote server is using to connect to local, if local is core. Empty otherwise
	UsingKey types.WeblensApiKey `json:"-" bson:"usingKey"`

	// Core or Backup
	Role types.ServerRole `json:"role" bson:"serverRole"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" bson:"isThisServer"`

	// Address of the remote server, only if the remote is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	CoreAddress string `json:"coreAddress" bson:"coreAddress"`

	UserCount int `json:"userCount" bson:"-"`
}

type WeblensDB struct {
	mongo    *mongo.Database
	useRedis bool
	redis    *redis.Client
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

const (
	FileShare  types.ShareType = "file"
	AlbumShare types.ShareType = "album"
)

const (
	FileGet types.RequestMode = "fileGet"

	// Grant access unconditionally. This is for sending
	// out updates where the user has already subscribed
	// elsewhere, and we just need to format the data for them
	WebsocketFileUpdate types.RequestMode = "wsFileUpdate"
	MarshalFile         types.RequestMode = "marshalFile"

	FileSubscribeRequest types.RequestMode = "fileSub"

	ApiKeyCreate types.RequestMode = "apiKeyCreate"
	ApiKeyGet    types.RequestMode = "apiKeyGet"

	BackupFileScan types.RequestMode = "backupFileScan"
)

type trashEntry struct {
	OrigParent   types.FileId `bson:"originalParentId"`
	OrigFilename string       `bson:"originalFilename"`
	TrashFileId  types.FileId `bson:"trashFileId"`
}

type AlbumData struct {
	Id             types.AlbumId     `bson:"_id"`
	Name           string            `bson:"name"`
	Owner          types.Username    `bson:"owner"`
	Cover          types.ContentId   `bson:"cover"`
	PrimaryColor   string            `bson:"primaryColor"`
	SecondaryColor string            `bson:"secondaryColor"`
	Medias         []types.ContentId `bson:"medias"`
	SharedWith     []types.Username  `bson:"sharedWith"`
	ShowOnTimeline bool              `bson:"showOnTimeline"`
}

type ApiKeyInfo struct {
	Id          primitive.ObjectID  `bson:"_id"`
	Key         types.WeblensApiKey `bson:"key"`
	Owner       types.Username      `bson:"owner"`
	CreatedTime time.Time           `bson:"createdTime"`
	RemoteUsing string              `bson:"remoteUsing"`
}

// type JournalResp struct {
// 	Journal []*fileJournalEntry `json:"journal"`
// }

// Errors

type WeblensUserError interface {
	types.WeblensError
}

var ErrNotUsingRedis = errors.New("not using redis")
var ErrNoCache = types.NewWeblensError("media references cache file that does not exist")

var ErrNoUser = types.NewWeblensError("user does not exist")
var ErrUserAlreadyExists = types.NewWeblensError("cannot create two users with the same username")
var ErrUserNotAuthorized = types.NewWeblensError("user does not have access the requested resource")
var ErrUserNotAuthenticated = types.NewWeblensError("user credentials are invalid")
var ErrNoFileAccess = types.NewWeblensError("user does not have access to file")
var ErrBadPassword = types.NewWeblensError("password provided does not authenticate user")

var ErrBadRequestMode = errors.New("access struct does not have correct request mode set for the given function")

var ErrNoMedia = errors.New("no media found")
var ErrNoImage = errors.New("media is missing required image")

var ErrNoShare = errors.New("no share found")
var ErrBadShareType = errors.New("expected share type does not match given share type")

var ErrUnsupportedImgType error = errors.New("image type is not supported by weblens")
var ErrPageOutOfRange = errors.New("page number does not exist on media")

var ErrNoKey = errors.New("api key is does not exist")
var ErrKeyInUse = errors.New("api key is already being used to identify another remote server")

var ErrAlreadyCore = errors.New("core server cannot have a remote core")
var ErrNotCore = errors.New("tried to perform core only action on non-core server")
var ErrNotBackup = errors.New("tried to perform backup only action on non-backup server")
var ErrAlreadyInit = errors.New("server is already initialized, cannot re-initialize server")

var ErrNoBackup = errors.New("no prior backups exist")
var ErrBadJournalAction = errors.New("unknown journal action type")
var ErrNoFile = types.NewWeblensError("file does not exist")
