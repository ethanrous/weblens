package http

import (
	"encoding/json"
	"fmt"

	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
)

// Websocket

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
	Key        types.SubId    `json:"subscribeKey"`
	TaskType   types.TaskType `json:"taskType"`
	LookingFor []string       `json:"lookingFor"`

	realKey types.SubId
}

func (tsm *taskSubscribeMeta) Action() types.WsAction {
	return types.TaskSubscribe
}

func (tsm *taskSubscribeMeta) GetKey() types.SubId {
	if tsm.realKey == "" {
		if tsm.Key != "" {
			tsm.realKey = types.SubId(fmt.Sprintf("TID#%s", tsm.Key))
		} else if tsm.TaskType != "" {
			tsm.realKey = types.SubId(fmt.Sprintf("TT#%s", tsm.TaskType))
		}
	}

	// if tsm.Key == "" && tsm.TaskType != "" {
	// 	taskPool := types.SERV.WorkerPool.GetTaskPoolByTaskType(tsm.TaskType)
	// 	if taskPool == nil {
	// 		return ""
	// 	}
	// 	tsm.Key = types.SubId(taskPool.ID())
	// }
	return tsm.realKey
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

type scanDirectoryMeta struct {
	Key types.SubId `json:"folderId"`
}

func (um *scanDirectoryMeta) Action() types.WsAction {
	return types.ScanDirectory
}

func (um *scanDirectoryMeta) GetKey() types.SubId {
	return um.Key
}

type cancelTaskMeta struct {
	TaskPoolId types.SubId `json:"taskPoolId"`
}

func (ctm *cancelTaskMeta) Action() types.WsAction {
	return types.CancelTask
}

func (ctm *cancelTaskMeta) GetKey() types.SubId {
	return ctm.TaskPoolId
}

// newActionBody returns a structure to hold the correct version of the websocket request body
func newActionBody(msg types.WsRequestInfo) (types.WsR, error) {
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
	case types.ScanDirectory:
		target := &scanDirectoryMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	case types.CancelTask:
		target := &cancelTaskMeta{}
		err := json.Unmarshal([]byte(msg.Content), target)
		return target, err
	default:
		return nil, werror.NewWeblensError(fmt.Sprint("did not recognize websocket action type: ", msg.Action))
	}
}

func (tsm taskSubscribeMeta) ResultKeys() []string {
	return tsm.LookingFor
}

// Physical of broadcasters to inform clients of updates in real time
