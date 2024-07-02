package dataStore

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WeblensDB struct {
	mongo *mongo.Database
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
	FileGet types.RequestMode = "fileGet"

	// WebsocketFileUpdate Grant access unconditionally. This is for sending
	// out updates where the user has already subscribed
	// elsewhere, and we just need to format the data for them
	WebsocketFileUpdate types.RequestMode = "wsFileUpdate"
	MarshalFile         types.RequestMode = "marshalFile"

	FileSubscribeRequest types.RequestMode = "fileSub"

	ApiKeyCreate types.RequestMode = "apiKeyCreate"
	ApiKeyGet    types.RequestMode = "apiKeyGet"

	BackupFileScan types.RequestMode = "backupFileScan"
)

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

var ErrNoCache = types.NewWeblensError("media references cache file that does not exist")

var ErrNoUser = types.NewWeblensError("user does not exist")

var ErrNoFileAccess = types.NewWeblensError("user does not have access to file")

var ErrBadRequestMode = errors.New("access struct does not have correct request mode set for the given function")

var ErrNoImage = errors.New("media is missing required image")

var ErrNoShare = errors.New("no share found")
var ErrBadShareType = errors.New("expected share type does not match given share type")

var ErrUnsupportedImgType = errors.New("image type is not supported by weblens")
var ErrPageOutOfRange = errors.New("page number does not exist on media")

var ErrNoKey = errors.New("api key is does not exist")
var ErrKeyInUse = errors.New("api key is already being used to identify another remote server")

var ErrNotCore = errors.New("tried to perform core only action on non-core server")
var ErrNotBackup = errors.New("tried to perform backup only action on non-backup server")
var ErrAlreadyInit = errors.New("server is already initialized, cannot re-initialize server")

var ErrNoBackup = errors.New("no prior backups exist")
var ErrBadJournalAction = errors.New("unknown journal action type")
var ErrNoFile = types.NewWeblensError("file does not exist")
