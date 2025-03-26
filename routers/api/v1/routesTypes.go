package http

import (
	"encoding/json"
	"fmt"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
)

type folderSubscribeMeta struct {
	Key       models.SubId   `json:"subscribeKey"`
	ShareId   models.ShareId `json:"shareId"`
	Recursive bool           `json:"recursive"`
}

func (fsm *folderSubscribeMeta) Action() models.WsAction {
	return models.FolderSubscribe
}

func (fsm *folderSubscribeMeta) GetKey() models.SubId {
	return fsm.Key
}

func (fsm *folderSubscribeMeta) GetShare(shareService models.ShareService) *models.FileShare {
	sh := shareService.Get(fsm.ShareId)
	if sh == nil {
		return nil
	}

	fileShare := sh.(*models.FileShare)
	return fileShare
}

type taskSubscribeMeta struct {
	Key     models.SubId `json:"subscribeKey"`
	JobName string       `json:"taskType"`

	realKey    models.SubId
	LookingFor []string `json:"lookingFor"`
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

func (tsm *taskSubscribeMeta) GetShare(shareService models.ShareService) *models.FileShare {
	return nil
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

func (um *unsubscribeMeta) GetShare(shareService models.ShareService) *models.FileShare {
	return nil
}

type scanDirectoryMeta struct {
	Key     models.SubId   `json:"folderId"`
	ShareId models.ShareId `json:"shareId"`
}

func (sdm *scanDirectoryMeta) Action() models.WsAction {
	return models.ScanDirectory
}

func (sdm *scanDirectoryMeta) GetKey() models.SubId {
	return sdm.Key
}

func (sdm *scanDirectoryMeta) GetShare(shareService models.ShareService) *models.FileShare {
	sh := shareService.Get(sdm.ShareId)
	if sh == nil {
		return nil
	}

	fileShare := sh.(*models.FileShare)
	return fileShare
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

func (ctm *cancelTaskMeta) GetShare(shareService models.ShareService) *models.FileShare {
	return nil
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg models.WsRequestInfo) (models.WsR, error) {
	switch msg.Action {
	case models.FolderSubscribe:
		target := &folderSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, werror.WithStack(err)
	case models.TaskSubscribe:
		target := &taskSubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, werror.WithStack(err)
	case models.Unsubscribe:
		target := &unsubscribeMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, werror.WithStack(err)
	case models.ScanDirectory:
		target := &scanDirectoryMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, werror.WithStack(err)
	case models.CancelTask:
		target := &cancelTaskMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, werror.WithStack(err)
	default:
		return nil, werror.Errorf("did not recognize websocket action type [%s]", msg.Action)
	}
}

func (tsm taskSubscribeMeta) ResultKeys() []string {
	return tsm.LookingFor
}
