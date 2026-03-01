// Package notify provides websocket notification services for broadcasting events to connected clients.
package notify

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/modules/wlstructs"
)

// Jobber is an interface representing a job or task with an ID and job name.
type Jobber interface {
	ID() string
	JobName() string
	GetStartTime() time.Time
}

// NewTaskNotification creates a websocket notification for a task event with the given result.
func NewTaskNotification(task Jobber, event websocket_mod.WsEvent, result task.Result) websocket_mod.WsResponseInfo {
	msg := websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    task.ID(),
		Content:         result.ToMap(),
		TaskType:        task.JobName(),
		TaskStartTime:   task.GetStartTime().UnixMilli(),
		BroadcastType:   websocket_mod.TaskSubscribe,
		ConstructedTime: time.Now().UnixMilli(),

		Sent: make(chan struct{}),
	}

	return msg
}

// NewPoolNotification creates a websocket notification for a task pool event with the given result.
func NewPoolNotification(pool *task.Pool, event websocket_mod.WsEvent, result task.Result) websocket_mod.WsResponseInfo {
	if pool.IsGlobal() {
		wlog.GlobalLogger().Warn().Msg("Not pushing update on global pool")

		return websocket_mod.WsResponseInfo{}
	}

	parentTask := pool.CreatedInTask()

	msg := websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    parentTask.ID(),
		Content:         result.ToMap(),
		TaskType:        parentTask.JobName(),
		BroadcastType:   websocket_mod.TaskSubscribe,
		ConstructedTime: time.Now().UnixMilli(),
	}

	return msg
}

// NewSystemNotification creates a websocket notification for a system-wide event with the given data.
func NewSystemNotification(event websocket_mod.WsEvent, data websocket_mod.WsData) websocket_mod.WsResponseInfo {
	msg := websocket_mod.WsResponseInfo{
		SubscribeKey:    websocket_mod.SystemSubscriberKey,
		EventTag:        event,
		Content:         data,
		BroadcastType:   websocket_mod.SystemSubscribe,
		ConstructedTime: time.Now().UnixMilli(),
	}

	return msg
}

// NewFileNotification creates websocket notifications for a file event, including notifications for the file,
// its parent folder, and optionally a pre-move parent if the file was moved.
func NewFileNotification(
	ctx context.Context,
	fileInfo wlstructs.FileInfo,
	event websocket_mod.WsEvent,
	options ...FileNotificationOptions,
) []websocket_mod.WsResponseInfo {
	if fileInfo.ID == "" {
		err := wlerrors.Errorf("File ID is empty")
		wlog.FromContext(ctx).Error().Stack().Err(err).Msg("Failed to create new file notification")

		return []websocket_mod.WsResponseInfo{}
	}

	o := consolidateFileOptions(options...)

	content := websocket_mod.WsData{"fileInfo": fileInfo}
	if o.MediaInfo.ContentID != "" {
		content["mediaData"] = options[0].MediaInfo
	}

	notifs := []websocket_mod.WsResponseInfo{}

	notifs = append(notifs, websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    fileInfo.ID,
		Content:         content,
		BroadcastType:   websocket_mod.FolderSubscribe,
		ConstructedTime: time.Now().UnixMilli(),
	})

	if fileInfo.ParentID != "" {
		parentMsg := notifs[0]
		parentMsg.SubscribeKey = fileInfo.ParentID
		notifs = append(notifs, parentMsg)
	}

	if o.PreMoveParentID != "" {
		preMoveMsg := notifs[0]
		preMoveMsg.SubscribeKey = o.PreMoveParentID
		notifs = append(notifs, preMoveMsg)
	}

	return notifs
}
