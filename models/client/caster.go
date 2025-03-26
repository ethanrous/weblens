package client

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/rest"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SimpleCaster struct {
	// cm        ClientManager
	msgChan   chan WsResponseInfo
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
		panic(werror.Errorf("Caster double close"))
	} else if c.global.Load() {
		return
	}
	c.enabled.Store(false)
	c.msgChan <- WsResponseInfo{}
}

func NewSimpleCaster(log *zerolog.Logger) *SimpleCaster {
	newCaster := &SimpleCaster{
		msgChan: make(chan WsResponseInfo, 100),
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

func (c *SimpleCaster) addToQueue(msg WsResponseInfo) {
	c.log.Trace().Func(func(e *zerolog.Event) {
		e.Str("websocket_event", string(msg.EventTag)).Msg("Caster adding message to queue")
	})
	c.flushLock.RLock()
	defer c.flushLock.RUnlock()
	c.msgChan <- (msg)
}

func (c *SimpleCaster) PushWeblensEvent(eventTag WsEvent, content ...WsC) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
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

func (c *SimpleCaster) PushTaskUpdate(task *task.Task, event WsEvent, result task.TaskResult) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.TaskId(),
		Content:       result.ToMap(),
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushPoolUpdate(
	pool task.Pool, event WsEvent, result task.TaskResult,
) {
	if !c.enabled.Load() {
		return
	}

	if pool.IsGlobal() {
		log.Warn().Msg("Not pushing update on global pool")
		return
	}

	parentTask := pool.CreatedInTask()

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  parentTask.TaskId(),
		Content:       result.ToMap(),
		TaskType:      parentTask.JobName(),
		BroadcastType: TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	// c.c.cm.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushShareUpdate(username string, newShareInfo share_model.FileShare) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      ShareUpdatedEvent,
		SubscribeKey:  username,
		Content:       WsC{"newShareInfo": newShareInfo},
		BroadcastType: UserSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileCreate(newFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:     FileCreatedEvent,
		SubscribeKey: newFile.GetParentId(),
		Content:      WsC{"fileInfo": newFile},

		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *media_model.Media) {
	if !c.enabled.Load() {
		return
	}

	mediaInfo := rest.MediaInfo{}
	if media != nil {
		mediaInfo = rest.MediaToMediaInfo(media)
	}

	msg := WsResponseInfo{
		EventTag:      FileUpdatedEvent,
		SubscribeKey:  updatedFile.ID(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": mediaInfo},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)

	if updatedFile.GetParent() == nil || updatedFile.GetParent().ID() == "ROOT" {
		return
	}

	msg = WsResponseInfo{
		EventTag:      FileUpdatedEvent,
		SubscribeKey:  updatedFile.GetParentId(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": mediaInfo},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesUpdate(updatedFiles []*fileTree.WeblensFileImpl, medias []*media_model.Media) {
	if !c.enabled.Load() {
		return
	}

	if len(updatedFiles) == 0 {
		return
	}

	mediaInfos := make([]rest.MediaInfo, 0, len(medias))
	for _, m := range medias {
		mediaInfos = append(mediaInfos, rest.MediaToMediaInfo(m))
	}

	msg := WsResponseInfo{
		EventTag:      FilesUpdatedEvent,
		SubscribeKey:  updatedFiles[0].GetParentId(),
		Content:       WsC{"filesInfo": updatedFiles, "mediaDatas": mediaInfos},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      FileMovedEvent,
		SubscribeKey:  preMoveFile.GetParentId(),
		Content:       WsC{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = WsResponseInfo{
		EventTag:      FileMovedEvent,
		SubscribeKey:  postMoveFile.GetParentId(),
		Content:       WsC{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      FilesMovedEvent,
		SubscribeKey:  preMoveParentId,
		Content:       WsC{"filesInfo": files},
		Error:         "",
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = WsResponseInfo{
		EventTag:      FilesMovedEvent,
		SubscribeKey:  postMoveParentId,
		Content:       WsC{"filesInfo": files},
		Error:         "",
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileDelete(deletedFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      FileDeletedEvent,
		SubscribeKey:  deletedFile.GetParent().ID(),
		Content:       WsC{"fileId": deletedFile.ID()},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesDelete(deletedFiles []*fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	fileIds := make(map[string][]fileTree.FileId)
	for _, f := range deletedFiles {
		list := fileIds[f.GetParentId()]
		if list == nil {
			fileIds[f.GetParentId()] = []fileTree.FileId{f.ID()}
		} else {
			fileIds[f.GetParentId()] = append(list, f.ID())
		}
	}

	for parentId, ids := range fileIds {
		msg := WsResponseInfo{
			EventTag:      FilesDeletedEvent,
			SubscribeKey:  parentId,
			Content:       WsC{"fileIds": ids},
			BroadcastType: FolderSubscribe,
			SentTime:      time.Now().Unix(),
		}

		c.addToQueue(msg)
	}

}

func (c *SimpleCaster) FolderSubToTask(folder fileTree.FileId, taskId task.Id) {
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

func (c *SimpleCaster) Relay(msg WsResponseInfo) {
	if !c.enabled.Load() {
		return
	}

	c.addToQueue(msg)
}
