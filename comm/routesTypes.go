package comm

import (
	"encoding/json"
	"fmt"

	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

type folderSubscribeMeta struct {
	Key       SubId          `json:"subscribeKey"`
	Recursive bool           `json:"recursive"`
	ShareId   models.ShareId `json:"shareId"`
}

func (fsm *folderSubscribeMeta) Action() WsAction {
	return FolderSubscribe
}

func (fsm *folderSubscribeMeta) GetKey() SubId {
	return fsm.Key
}

type taskSubscribeMeta struct {
	Key        SubId    `json:"subscribeKey"`
	JobName    string   `json:"taskType"`
	LookingFor []string `json:"lookingFor"`

	realKey SubId
}

func (tsm *taskSubscribeMeta) Action() WsAction {
	return TaskSubscribe
}

func (tsm *taskSubscribeMeta) GetKey() SubId {
	if tsm.realKey == "" {
		if tsm.Key != "" {
			tsm.realKey = SubId(fmt.Sprintf("TID#%s", tsm.Key))
		} else if tsm.JobName != "" {
			tsm.realKey = SubId(fmt.Sprintf("TT#%s", tsm.JobName))
		}
	}
	return tsm.realKey
}

type unsubscribeMeta struct {
	Key SubId `json:"subscribeKey"`
}

func (um *unsubscribeMeta) Action() WsAction {
	return Unsubscribe
}

func (um *unsubscribeMeta) GetKey() SubId {
	return um.Key
}

type scanDirectoryMeta struct {
	Key SubId `json:"folderId"`
}

func (um *scanDirectoryMeta) Action() WsAction {
	return ScanDirectory
}

func (um *scanDirectoryMeta) GetKey() SubId {
	return um.Key
}

type cancelTaskMeta struct {
	TaskPoolId SubId `json:"taskPoolId"`
}

func (ctm *cancelTaskMeta) Action() WsAction {
	return CancelTask
}

func (ctm *cancelTaskMeta) GetKey() SubId {
	return ctm.TaskPoolId
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg WsRequestInfo) (WsR, error) {
	switch msg.Action {
	case FolderSubscribe:
		target := &folderSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case TaskSubscribe:
		target := &taskSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case Unsubscribe:
		target := &unsubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case ScanDirectory:
		target := &scanDirectoryMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case CancelTask:
		target := &cancelTaskMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	default:
		return nil, werror.Errorf("did not recognize websocket action type [%s]", msg.Action)
	}
}

func (tsm taskSubscribeMeta) ResultKeys() []string {
	return tsm.LookingFor
}
