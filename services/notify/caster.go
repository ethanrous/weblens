package notify

import (
	"context"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	task_model "github.com/ethanrous/weblens/models/task"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewTaskNotification(task *task_model.Task, event websocket_mod.WsEvent, result task_mod.TaskResult) websocket_mod.WsResponseInfo {
	msg := websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    task.Id(),
		Content:         result.ToMap(),
		TaskType:        task.JobName(),
		BroadcastType:   websocket_mod.TaskSubscribe,
		ConstructedTime: time.Now().Unix(),
	}

	return msg
}

func NewPoolNotification(pool task_mod.Pool, event websocket_mod.WsEvent, result task_mod.TaskResult) websocket_mod.WsResponseInfo {
	if pool.IsGlobal() {
		log.Warn().Msg("Not pushing update on global pool")

		return websocket_mod.WsResponseInfo{}
	}

	parentTask := pool.CreatedInTask()

	msg := websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    parentTask.Id(),
		Content:         result.ToMap(),
		TaskType:        parentTask.JobName(),
		BroadcastType:   websocket_mod.TaskSubscribe,
		ConstructedTime: time.Now().Unix(),
	}

	return msg
}

func NewSystemNotification(event websocket_mod.WsEvent, data websocket_mod.WsData) websocket_mod.WsResponseInfo {
	msg := websocket_mod.WsResponseInfo{
		SubscribeKey:    websocket_mod.SystemSubscriberKey,
		EventTag:        event,
		Content:         data,
		BroadcastType:   websocket_mod.SystemSubscribe,
		ConstructedTime: time.Now().Unix(),
	}

	return msg
}

func NewFileNotification(
	c context.Context,
	file *file_model.WeblensFileImpl,
	event websocket_mod.WsEvent,
	options ...FileNotificationOptions,
) []websocket_mod.WsResponseInfo {
	ctx, _ := context_service.FromContext(c)

	fileInfo, err := reshape.WeblensFileToFileInfo(ctx, file, false)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to create new file notification")

		return []websocket_mod.WsResponseInfo{}
	}

	if file.ID() == "" {
		err = errors.Errorf("File ID is empty")
		ctx.Log().Error().Stack().Err(err).Msg("Failed to create new file notification")

		return []websocket_mod.WsResponseInfo{}
	}

	o := consolidateFileOptions(options...)

	content := websocket_mod.WsData{"fileInfo": fileInfo}
	if o.MediaInfo.ContentId != "" {
		content["mediaData"] = options[0].MediaInfo
	}

	notifs := []websocket_mod.WsResponseInfo{}

	notifs = append(notifs, websocket_mod.WsResponseInfo{
		EventTag:        event,
		SubscribeKey:    file.ID(),
		Content:         content,
		BroadcastType:   websocket_mod.FolderSubscribe,
		ConstructedTime: time.Now().Unix(),
	})

	if file.GetParent() != nil && !file.GetParent().GetPortablePath().IsRoot() {
		parentMsg := notifs[0]
		parentMsg.SubscribeKey = file.GetParent().ID()
		notifs = append(notifs, parentMsg)
	}

	if o.PreMoveParentId != "" {
		preMoveMsg := notifs[0]
		preMoveMsg.SubscribeKey = o.PreMoveParentId
		notifs = append(notifs, preMoveMsg)
	}

	return notifs
}
