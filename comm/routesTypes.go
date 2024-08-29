package comm

import (
	"encoding/json"
	"fmt"

	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

type folderSubscribeMeta struct {
	Key       models.SubId `json:"subscribeKey"`
	Recursive bool         `json:"recursive"`
	ShareId   models.ShareId `json:"shareId"`
}

func (fsm *folderSubscribeMeta) Action() models.WsAction {
	return models.FolderSubscribe
}

func (fsm *folderSubscribeMeta) GetKey() models.SubId {
	return fsm.Key
}

type taskSubscribeMeta struct {
	Key     models.SubId `json:"subscribeKey"`
	JobName string       `json:"taskType"`
	LookingFor []string `json:"lookingFor"`

	realKey models.SubId
}

func (tsm *taskSubscribeMeta) Action() models.WsAction {
	return models.TaskSubscribe
}

func (tsm *taskSubscribeMeta) GetKey() models.SubId {
	if tsm.realKey == "" {
		if tsm.Key != "" {
			tsm.realKey = models.SubId(fmt.Sprintf("TID#%s", tsm.Key))
		} else if tsm.JobName != "" {
			tsm.realKey = models.SubId(fmt.Sprintf("TT#%s", tsm.JobName))
		}
	}
	return tsm.realKey
}

type unsubscribeMeta struct {
	Key models.SubId `json:"subscribeKey"`
}

func (um *unsubscribeMeta) Action() models.WsAction {
	return models.Unsubscribe
}

func (um *unsubscribeMeta) GetKey() models.SubId {
	return um.Key
}

type scanDirectoryMeta struct {
	Key models.SubId `json:"folderId"`
}

func (um *scanDirectoryMeta) Action() models.WsAction {
	return models.ScanDirectory
}

func (um *scanDirectoryMeta) GetKey() models.SubId {
	return um.Key
}

type cancelTaskMeta struct {
	TaskPoolId models.SubId `json:"taskPoolId"`
}

func (ctm *cancelTaskMeta) Action() models.WsAction {
	return models.CancelTask
}

func (ctm *cancelTaskMeta) GetKey() models.SubId {
	return ctm.TaskPoolId
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg models.WsRequestInfo) (models.WsR, error) {
	switch msg.Action {
	case models.FolderSubscribe:
		target := &folderSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case models.TaskSubscribe:
		target := &taskSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case models.Unsubscribe:
		target := &unsubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case models.ScanDirectory:
		target := &scanDirectoryMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case models.CancelTask:
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
