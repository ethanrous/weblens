package notify

import (
	"sync"
	"sync/atomic"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/structs"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SimpleCaster struct {
	// cm        ClientManager
	msgChan   chan websocket_mod.WsResponseInfo
	flushLock sync.RWMutex
	enabled   atomic.Bool
	global    atomic.Bool
	log       *zerolog.Logger
}

func (c *SimpleCaster) DisableAutoFlush() {
	// no-op
}

func (c *SimpleCaster) AutoFlushEnable() {
	// no-op
}

func (c *SimpleCaster) Flush() {
	log.Trace().Msg("Caster flushing message queue")
	c.flushLock.Lock()
	for len(c.msgChan) != 0 {
		log.Trace().Msg("Caster waiting for message queue to be empty...")
		time.Sleep(10 * time.Millisecond)
	}

	log.Trace().Msg("Caster flush complete")
	c.flushLock.Unlock()
}

func (c *SimpleCaster) Global() {
	c.global.Store(true)
}

func (c *SimpleCaster) Close() {
	if !c.enabled.Load() {
		panic(errors.Errorf("Caster double close"))
	} else if c.global.Load() {
		return
	}
	c.enabled.Store(false)
	c.msgChan <- websocket_mod.WsResponseInfo{}
}

func NewSimpleCaster(log *zerolog.Logger) *SimpleCaster {
	newCaster := &SimpleCaster{
		msgChan: make(chan websocket_mod.WsResponseInfo, 100),
		log:     log,
	}

	newCaster.enabled.Store(true)

	// go newCaster.msgWorker(cm)

	return newCaster
}

func (c *SimpleCaster) Enable() {
	c.enabled.Store(true)
}

func (c *SimpleCaster) Disable() {
	c.enabled.Store(false)
}

func (c *SimpleCaster) IsBuffered() bool {
	return false
}

func (c *SimpleCaster) IsEnabled() bool {
	return c.enabled.Load()
}

// func (c *SimpleCaster) msgWorker(cm ClientManager) {
// 	for msg := range c.msgChan {
// 		if !c.enabled.Load() && msg.EventTag == "" {
// 			break
// 		}
//
// 		cm.Send(msg)
// 	}
//
// 	log.Trace().Msg("Caster message worker exiting")
//
// 	close(c.msgChan)
// }

func (c *SimpleCaster) addToQueue(msg websocket_mod.WsResponseInfo) {
	c.log.Trace().Func(func(e *zerolog.Event) {
		e.Str("websocket_event", string(msg.EventTag)).Msg("Caster adding message to queue")
	})
	c.flushLock.RLock()
	defer c.flushLock.RUnlock()
	c.msgChan <- (msg)
}

func (c *SimpleCaster) PushWeblensEvent(eventTag websocket_mod.WsEvent, content ...websocket_mod.WsData) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  "WEBLENS",
		BroadcastType: "serverEvent",
		SentTime:      time.Now().Unix(),
	}

	if len(content) != 0 {
		msg.Content = content[0]
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushTaskUpdate(task *task_model.Task, event websocket_mod.WsEvent, result task_mod.TaskResult) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.Id(),
		Content:       result.ToMap(),
		TaskType:      task.JobName(),
		BroadcastType: websocket_mod.TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

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
		EventTag:      event,
		SubscribeKey:  parentTask.Id(),
		Content:       result.ToMap(),
		TaskType:      parentTask.JobName(),
		BroadcastType: websocket_mod.TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	return msg
}

func NewSystemNotification(event websocket_mod.WsEvent, data websocket_mod.WsData) websocket_mod.WsResponseInfo {
	msg := websocket_mod.WsResponseInfo{
		EventTag:        event,
		Content:         data,
		BroadcastType:   "serverEvent",
		ConstructedTime: time.Now().Unix(),
	}

	return msg
}

func NewFileNotification(ctx context.ContextZ, file *file_model.WeblensFileImpl, event websocket_mod.WsEvent, mediaInfo structs.MediaInfo) []websocket_mod.WsResponseInfo {
	fileInfo, err := reshape.WeblensFileToFileInfo(ctx, file, false)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to create new file notification")
		return []websocket_mod.WsResponseInfo{}
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FileUpdatedEvent,
		SubscribeKey:  file.ID(),
		Content:       websocket_mod.WsData{"fileInfo": fileInfo, "mediaData": mediaInfo},
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	if file.GetParent() == nil || file.GetParent().GetPortablePath().IsRoot() {
		return []websocket_mod.WsResponseInfo{msg}
	}

	parentMsg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FileUpdatedEvent,
		SubscribeKey:  file.GetParent().ID(),
		Content:       websocket_mod.WsData{"fileInfo": fileInfo, "mediaData": mediaInfo},
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	return []websocket_mod.WsResponseInfo{msg, parentMsg}
}

func (c *SimpleCaster) PushPoolUpdate(
	pool task_mod.Pool, event websocket_mod.WsEvent, result task_mod.TaskResult,
) {
	if !c.enabled.Load() {
		return
	}

	if pool.IsGlobal() {
		log.Warn().Msg("Not pushing update on global pool")
		return
	}

	parentTask := pool.CreatedInTask()

	msg := websocket_mod.WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  parentTask.Id(),
		Content:       result.ToMap(),
		TaskType:      parentTask.JobName(),
		BroadcastType: websocket_mod.TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	// c.c.cm.Send(string(event), types.SubId(taskId), []types.websocket_mod.WsData{types.websocket_mod.WsData(result)})
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushShareUpdate(username string, newShareInfo share_model.FileShare) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.ShareUpdatedEvent,
		SubscribeKey:  username,
		Content:       websocket_mod.WsData{"newShareInfo": newShareInfo},
		BroadcastType: websocket_mod.UserSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileCreate(newFile *file_model.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:     websocket_mod.FileCreatedEvent,
		SubscribeKey: newFile.GetParentId(),
		Content:      websocket_mod.WsData{"fileInfo": newFile},

		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

// func (c *SimpleCaster) PushFilesUpdate(updatedFiles []*file_model.WeblensFileImpl, medias []*media_model.Media) {
// 	if !c.enabled.Load() {
// 		return
// 	}
//
// 	if len(updatedFiles) == 0 {
// 		return
// 	}
//
// 	mediaInfos := make([]structs.MediaInfo, 0, len(medias))
// 	for _, m := range medias {
// 		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
// 	}
//
// 	msg := websocket_mod.WsResponseInfo{
// 		EventTag:      websocket_mod.FilesUpdatedEvent,
// 		SubscribeKey:  updatedFiles[0].GetParentId(),
// 		Content:       websocket_mod.WsData{"filesInfo": updatedFiles, "mediaDatas": mediaInfos},
// 		BroadcastType: websocket_mod.FolderSubscribe,
// 		SentTime:      time.Now().Unix(),
// 	}
//
// 	c.addToQueue(msg)
// }

func (c *SimpleCaster) PushFileMove(preMoveFile *file_model.WeblensFileImpl, postMoveFile *file_model.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FileMovedEvent,
		SubscribeKey:  preMoveFile.GetParentId(),
		Content:       websocket_mod.WsData{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FileMovedEvent,
		SubscribeKey:  postMoveFile.GetParentId(),
		Content:       websocket_mod.WsData{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesMove(preMoveParentId, postMoveParentId string, files []*file_model.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FilesMovedEvent,
		SubscribeKey:  preMoveParentId,
		Content:       websocket_mod.WsData{"filesInfo": files},
		Error:         "",
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FilesMovedEvent,
		SubscribeKey:  postMoveParentId,
		Content:       websocket_mod.WsData{"filesInfo": files},
		Error:         "",
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileDelete(deletedFile *file_model.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := websocket_mod.WsResponseInfo{
		EventTag:      websocket_mod.FileDeletedEvent,
		SubscribeKey:  deletedFile.GetParent().ID(),
		Content:       websocket_mod.WsData{"fileId": deletedFile.ID()},
		BroadcastType: websocket_mod.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesDelete(deletedFiles []*file_model.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	fileIds := make(map[string][]string)
	for _, f := range deletedFiles {
		list := fileIds[f.GetParentId()]
		if list == nil {
			fileIds[f.GetParentId()] = []string{f.ID()}
		} else {
			fileIds[f.GetParentId()] = append(list, f.ID())
		}
	}

	for parentId, ids := range fileIds {
		msg := websocket_mod.WsResponseInfo{
			EventTag:      websocket_mod.FilesDeletedEvent,
			SubscribeKey:  parentId,
			Content:       websocket_mod.WsData{"fileIds": ids},
			BroadcastType: websocket_mod.FolderSubscribe,
			SentTime:      time.Now().Unix(),
		}

		c.addToQueue(msg)
	}

}

func (c *SimpleCaster) FolderSubToTask(folder, taskId string) {
	panic("not implemented")

	// if !c.enabled.Load() {
	// 	return
	// }
	//
	// subs := c.cm.GetSubscribers(FolderSubscribe, folder)
	//
	// for _, s := range subs {
	// 	_, _, err := c.cm.Subscribe(s, taskId, TaskSubscribe, time.Now(), nil)
	// 	if err != nil {
	// 		c.log.Error().Stack().Err(err).Msg("")
	// 	}
	// }
}

func (c *SimpleCaster) Relay(msg websocket_mod.WsResponseInfo) {
	if !c.enabled.Load() {
		return
	}

	c.addToQueue(msg)
}
