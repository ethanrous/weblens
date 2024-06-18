package routes

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
)

// Websocket

type wsAuthorize struct {
	Auth string `json:"auth"`
}

type wsResponse struct {
	EventTag     string        `json:"eventTag"`
	SubscribeKey types.SubId   `json:"subscribeKey"`
	Content      []types.WsMsg `json:"content"`
	Error        string        `json:"error"`

	broadcastType types.WsAction
}

type wsRequest struct {
	Action  types.WsAction `json:"action"`
	Content string         `json:"content"`
}

type folderSubscribeMeta struct {
	Key       types.SubId   `json:"subscribeKey"`
	Recursive bool          `json:"recursive"`
	ShareId   types.ShareId `json:"shareId"`
}

func (fsm *folderSubscribeMeta) Action() types.WsAction {
	return types.FolderSubscribe
}

func (fsm *folderSubscribeMeta) GetKey() types.SubId {
	return fsm.Key
}

type taskSubscribeMeta struct {
	Key        types.SubId `json:"subscribeKey"`
	LookingFor []string    `json:"lookingFor"`
}

func (tsm *taskSubscribeMeta) Action() types.WsAction {
	return types.TaskSubscribe
}

func (tsm *taskSubscribeMeta) GetKey() types.SubId {
	return tsm.Key
}

type unsubscribeMeta struct {
	Key types.SubId `json:"subscribeKey"`
}

func (um *unsubscribeMeta) Action() types.WsAction {
	return types.Unsubscribe
}

func (um *unsubscribeMeta) GetKey() types.SubId {
	return um.Key
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg wsRequest) (types.WsContent, error) {
	switch msg.Action {
	case types.FolderSubscribe:
		target := &folderSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case types.TaskSubscribe:
		target := &taskSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case types.Unsubscribe:
		target := &unsubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	default:
		return nil, types.NewWeblensError(fmt.Sprint("did not recognize websocket action type: ", msg.Action))
	}
}

func (tsm taskSubscribeMeta) ResultKeys() []string {
	return tsm.LookingFor
}

type scanBody struct {
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
	mediaRepo         types.MediaRepo
}

type unbufferedCaster struct {
	enabled   bool
	mediaRepo types.MediaRepo
}

var ErrBadAuthScheme = types.NewWeblensError("invalid authorization scheme")
var ErrBasicAuthFormat = types.NewWeblensError("did not get expected encoded basic auth format")
var ErrEmptyAuth = types.NewWeblensError("empty auth header not allowed on endpoint")
var ErrCoreOriginate = types.NewWeblensError("core server attempted to ping remote server")
var ErrNoAddress = types.NewWeblensError("trying to make request to core without a core address")
var ErrNoKey = types.NewWeblensError("trying to make request to core without an api key")
var ErrNoBody = types.NewWeblensError("trying to read http body with no content")
var ErrBodyNotAllowed = types.NewWeblensError("trying to read http body of GET request")
var ErrCasterDoubleClose = types.NewWeblensError("trying to close an already disabled caster")
var ErrUnknownWebsocketAction = types.NewWeblensError("did not recognize websocket action type")
