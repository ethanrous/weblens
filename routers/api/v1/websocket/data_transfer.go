package websocket

import (
	"encoding/json"
	"fmt"

	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/pkg/errors"
)

type folderSubscribeMeta struct {
	Key       string `json:"subscribeKey"`
	ShareId   string `json:"shareId"`
	Recursive bool   `json:"recursive"`
}

func (fsm *folderSubscribeMeta) Action() websocket_mod.WsAction {
	return websocket_mod.FolderSubscribe
}

func (fsm *folderSubscribeMeta) GetKey() string {
	return fsm.Key
}

func (fsm *folderSubscribeMeta) GetShareId() string {
	return fsm.ShareId
}

type taskSubscribeMeta struct {
	Key     string `json:"subscribeKey"`
	JobName string `json:"taskType"`

	realKey    string
	LookingFor []string `json:"lookingFor"`
}

func (tsm *taskSubscribeMeta) Action() websocket_mod.WsAction {
	return websocket_mod.TaskSubscribe
}

func (tsm *taskSubscribeMeta) GetKey() string {
	if tsm.realKey == "" {
		if tsm.Key != "" {
			tsm.realKey = string(fmt.Sprintf("TID#%s", tsm.Key))
		} else if tsm.JobName != "" {
			tsm.realKey = string(fmt.Sprintf("TT#%s", tsm.JobName))
		}
	}
	return tsm.realKey
}
func (tsm *taskSubscribeMeta) GetShareId() string {
	return ""
}

type unsubscribeMeta struct {
	Key string `json:"subscribeKey"`
}

func (um *unsubscribeMeta) Action() websocket_mod.WsAction {
	return websocket_mod.Unsubscribe
}

func (um *unsubscribeMeta) GetKey() string {
	return um.Key
}
func (um *unsubscribeMeta) GetShareId() string {
	return ""
}

type scanDirectoryMeta struct {
	Key     string `json:"folderId"`
	ShareId string `json:"shareId"`
}

func (sdm *scanDirectoryMeta) Action() websocket_mod.WsAction {
	return websocket_mod.ScanDirectory
}

func (sdm *scanDirectoryMeta) GetKey() string {
	return sdm.Key
}
func (sdm *scanDirectoryMeta) GetShareId() string {
	return sdm.ShareId
}

type cancelTaskMeta struct {
	TaskPoolId string `json:"taskPoolId"`
}

func (ctm *cancelTaskMeta) Action() websocket_mod.WsAction {
	return websocket_mod.CancelTask
}

func (ctm *cancelTaskMeta) GetKey() string {
	return ctm.TaskPoolId
}
func (ctm *cancelTaskMeta) GetShareId() string {
	return ""
}

func (tsm taskSubscribeMeta) ResultKeys() []string {
	return tsm.LookingFor
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg websocket_mod.WsRequestInfo) (websocket_mod.WsR, error) {
	switch msg.Action {
	case websocket_mod.FolderSubscribe:
		target := &folderSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, errors.WithStack(err)
	case websocket_mod.TaskSubscribe:
		target := &taskSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, errors.WithStack(err)
	case websocket_mod.Unsubscribe:
		target := &unsubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, errors.WithStack(err)
	case websocket_mod.ScanDirectory:
		target := &scanDirectoryMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, errors.WithStack(err)
	case websocket_mod.CancelTask:
		target := &cancelTaskMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, errors.WithStack(err)
	default:
		return nil, errors.Errorf("did not recognize websocket action type [%s]", msg.Action)
	}
}
