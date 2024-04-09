package routes

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gorilla/websocket"
)

// Endpoint logic

type loginInfo struct {
	Username types.Username `json:"username"`
	Password string         `json:"password"`
}

type fileUpdateInfo struct {
	NewName     string       `json:"newName"`
	NewParentId types.FileId `json:"newParentId"`
}

type updateMany struct {
	Files       []types.FileId `json:"fileIds"`
	NewParentId types.FileId   `json:"newParentId"`
}

type takeoutFiles struct {
	FileIds []types.FileId `json:"fileIds"`
}

type tokenReturn struct {
	Token string `json:"token"`
}

type newUserInfo struct {
	Username     types.Username `json:"username"`
	Password     string         `json:"password"`
	Admin        bool           `json:"admin"`
	AutoActivate bool           `json:"autoActivate"`
}

type newFileInfo struct {
	ParentFolderId types.FileId `json:"parentFolderId"`
	NewFileName    string       `json:"newFileName"`
	FileSize       int64        `json:"fileSize"`
}

type newUploadInfo struct {
	RootFolderId    types.FileId `json:"rootFolderId"`
	ChunkSize       int64        `json:"chunkSize"`
	TotalUploadSize int64        `json:"totalUploadSize"`
}

type passwordUpdateInfo struct {
	OldPass string `json:"oldPassword"`
	NewPass string `json:"newPassword"`
}

// type fileShare struct {
// 	Files []types.FileId   `json:"files"`
// 	Users []types.Username `json:"users"`
// }

type newShareInfo struct {
	FileIds  []types.FileId   `json:"fileIds"`
	Users    []types.Username `json:"users"`
	Public   bool             `json:"public"`
	Wormhole bool             `json:"wormhole"`
}

type initServer struct {
	Name string           `json:"name"`
	Role types.ServerRole `json:"role"`

	Username    types.Username `json:"username"`
	Password    string         `json:"password"`
	CoreAddress string         `json:"coreAddress"`
}

type newServerInfo struct {
	Id       string           `json:"serverId"`
	Role     types.ServerRole `json:"role"`
	Name     string           `json:"name"`
	UsingKey string           `json:"usingKey"`
}

// type deleteShareInfo struct {
// 	ShareId types.ShareId `json:"shareId"`
// }

// type userInfo struct {
// 	Username      types.Username `json:"username"`
// 	HomeFolderId  types.FileId   `json:"homeId"`
// 	TrashFolderId types.FileId   `json:"trashId"`
// 	Admin         bool           `json:"admin"`
// 	Activated     bool           `json:"activated"`
// }

// Websocket

type subType string
type subId string

type wsM map[string]any

type wsAuthorize struct {
	Auth string `json:"auth"`
}

type wsResponse struct {
	MessageStatus string `json:"messageStatus"`
	SubscribeKey  subId  `json:"subscribeKey"`
	Content       []wsM  `json:"content"`
	Error         string `json:"error"`

	broadcastType subType
}

type wsAction string

const (
	Subscribe     wsAction = "subscribe"
	Unsubscribe   wsAction = "unsubscribe"
	ScanDirectory wsAction = "scan_directory"
)

type wsRequest struct {
	Action  wsAction `json:"action"`
	Content string   `json:"content"`
}

type subMeta interface {
	Meta(subType) subMeta
}

type subscribeMetadata string

type subscribeInfo struct {
	SubType subType           `json:"subscribeType"`
	Key     subId             `json:"subscribeKey"`
	Meta    subscribeMetadata `json:"subscribeMeta"`
}

type unsubscribeInfo struct {
	Key subId `json:"subscribeKey"`
}

func (s subscribeMetadata) Meta(t subType) subMeta {
	var ret subMeta
	switch t {
	case SubTask:
		meta := taskSubMetadata{}
		err := json.Unmarshal([]byte(s), &meta)
		if err != nil {
			util.ErrTrace(err)
			return nil
		}
		ret = meta
	default:
		return nil
	}
	return ret
}

type taskSubMetadata struct {
	LookingFor []string `json:"lookingFor"`
}

func (task taskSubMetadata) Meta(t subType) subMeta {
	return task
}

func (task taskSubMetadata) ResultKeys() []string {
	return task.LookingFor
}

type scanInfo struct {
	FolderId  types.FileId `json:"folderId"`
	Filename  string       `json:"filename"`
	Recursive bool         `json:"recursive"`
	DeepScan  bool         `json:"full"`
}

// Physical of broadcasters to inform clients of updates in real time

type bufferedCaster struct {
	bufLimit          int
	buffer            []wsResponse
	autoFlush         bool
	enabled           bool
	autoFlushInterval time.Duration
	bufLock           *sync.Mutex
}

type unbufferedCaster struct {
	enabled    bool
	recipients []*Client
}

// var Caster *caster = &caster{enabled: false}
var Caster = NewBufferedCaster()

// Broadcaster that is always disabled
var VoidCaster *unbufferedCaster = &unbufferedCaster{enabled: false}

var UploadTasker types.TaskPool

// Client

const (
	SubFolder subType = "folder"
	SubTask   subType = "task"
	SubUser   subType = "user" // This one does not actually get "subscribed" to, it is automatically tracked for every websocket
)

type subscription struct {
	Type subType
	Key  subId
}

type clientId string

type Client struct {
	connId        clientId
	conn          *websocket.Conn
	mu            sync.Mutex
	subscriptions []subscription
	user          types.User
}

type clientManager struct {
	// Key: connection id, value: client instance
	clientMap map[clientId]*Client
	clientMu  *sync.Mutex

	// Key: subscription identifier, value: connection id
	// Use string -> bool map to take advantage of O(1) lookup time when removing clients
	// Bool represents if the subscription is `recursive`
	// {
	// 	"fileId": {
	// 		"clientId1": true
	// 		"clientId2": false
	// 	}
	// }
	folderSubs map[subId][]*Client
	taskSubs   map[subId][]*Client
	folderMu   *sync.Mutex
	taskMu     *sync.Mutex
}

var ErrBadAuthScheme types.WeblensError = errors.New("invalid authorization scheme")
var ErrBasicAuthFormat types.WeblensError = errors.New("did not get expected encoded basic auth format")
var ErrEmptyAuth types.WeblensError = errors.New("empty auth header not allowed on endpoint")
var ErrCoreOriginate types.WeblensError = errors.New("core server attempted to ping remote server")
