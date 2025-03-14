package caster

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var _ models.Broadcaster = (*SimpleCaster)(nil)

type SimpleCaster struct {
	cm        models.ClientManager
	msgChan   chan models.WsResponseInfo
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
	c.msgChan <- models.WsResponseInfo{}
}

func NewSimpleCaster(cm models.ClientManager, log *zerolog.Logger) *SimpleCaster {
	newCaster := &SimpleCaster{
		cm:      cm,
		msgChan: make(chan models.WsResponseInfo, 100),
		log:     log,
	}

	newCaster.enabled.Store(true)

	go newCaster.msgWorker(cm)

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

func (c *SimpleCaster) msgWorker(cm models.ClientManager) {
	for msg := range c.msgChan {
		if !c.enabled.Load() && msg.EventTag == "" {
			break
		}

		cm.Send(msg)
	}

	log.Trace().Msg("Caster message worker exiting")

	close(c.msgChan)
}

func (c *SimpleCaster) addToQueue(msg models.WsResponseInfo) {
	c.log.Trace().Func(func(e *zerolog.Event) { e.Str("websocket_event", msg.EventTag).Msg("Caster adding message to queue") })
	c.flushLock.RLock()
	defer c.flushLock.RUnlock()
	c.msgChan <- (msg)
}

func (c *SimpleCaster) PushWeblensEvent(eventTag string, content ...models.WsC) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
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

func (c *SimpleCaster) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.TaskId(),
		Content:       result.ToMap(),
		TaskType:      task.JobName(),
		BroadcastType: models.TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushPoolUpdate(
	pool task.Pool, event string, result task.TaskResult,
) {
	if !c.enabled.Load() {
		return
	}

	if pool.IsGlobal() {
		log.Warn().Msg("Not pushing update on global pool")
		return
	}

	parentTask := pool.CreatedInTask()

	msg := models.WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  parentTask.TaskId(),
		Content:       result.ToMap(),
		TaskType:      parentTask.JobName(),
		BroadcastType: models.TaskSubscribe,
		SentTime:      time.Now().Unix(),
	}

	// c.c.cm.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushShareUpdate(username models.Username, newShareInfo models.Share) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:      models.ShareUpdatedEvent,
		SubscribeKey:  username,
		Content:       models.WsC{"newShareInfo": newShareInfo},
		BroadcastType: models.UserSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileCreate(newFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:     models.FileCreatedEvent,
		SubscribeKey: newFile.GetParentId(),
		Content:      models.WsC{"fileInfo": newFile},

		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *models.Media) {
	if !c.enabled.Load() {
		return
	}

	mediaInfo := rest.MediaInfo{}
	if media != nil {
		mediaInfo = rest.MediaToMediaInfo(media)
	}

	msg := models.WsResponseInfo{
		EventTag:      models.FileUpdatedEvent,
		SubscribeKey:  updatedFile.ID(),
		Content:       models.WsC{"fileInfo": updatedFile, "mediaData": mediaInfo},
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)

	if updatedFile.GetParent() == nil || updatedFile.GetParent().ID() == "ROOT" {
		return
	}

	msg = models.WsResponseInfo{
		EventTag:      models.FileUpdatedEvent,
		SubscribeKey:  updatedFile.GetParentId(),
		Content:       models.WsC{"fileInfo": updatedFile, "mediaData": mediaInfo},
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesUpdate(updatedFiles []*fileTree.WeblensFileImpl, medias []*models.Media) {
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

	msg := models.WsResponseInfo{
		EventTag:      models.FilesUpdatedEvent,
		SubscribeKey:  updatedFiles[0].GetParentId(),
		Content:       models.WsC{"filesInfo": updatedFiles, "mediaDatas": mediaInfos},
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}

	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:      models.FileMovedEvent,
		SubscribeKey:  preMoveFile.GetParentId(),
		Content:       models.WsC{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = models.WsResponseInfo{
		EventTag:      models.FileMovedEvent,
		SubscribeKey:  postMoveFile.GetParentId(),
		Content:       models.WsC{"fileInfo": postMoveFile},
		Error:         "",
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:      models.FilesMovedEvent,
		SubscribeKey:  preMoveParentId,
		Content:       models.WsC{"filesInfo": files},
		Error:         "",
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)

	msg = models.WsResponseInfo{
		EventTag:      models.FilesMovedEvent,
		SubscribeKey:  postMoveParentId,
		Content:       models.WsC{"filesInfo": files},
		Error:         "",
		BroadcastType: models.FolderSubscribe,
		SentTime:      time.Now().Unix(),
	}
	c.addToQueue(msg)
}

func (c *SimpleCaster) PushFileDelete(deletedFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := models.WsResponseInfo{
		EventTag:      models.FileDeletedEvent,
		SubscribeKey:  deletedFile.GetParent().ID(),
		Content:       models.WsC{"fileId": deletedFile.ID()},
		BroadcastType: models.FolderSubscribe,
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
		msg := models.WsResponseInfo{
			EventTag:      models.FilesDeletedEvent,
			SubscribeKey:  parentId,
			Content:       models.WsC{"fileIds": ids},
			BroadcastType: models.FolderSubscribe,
			SentTime:      time.Now().Unix(),
		}

		c.addToQueue(msg)
	}

}

func (c *SimpleCaster) FolderSubToTask(folder fileTree.FileId, taskId task.Id) {
	if !c.enabled.Load() {
		return
	}

	subs := c.cm.GetSubscribers(models.FolderSubscribe, folder)

	for _, s := range subs {
		_, _, err := c.cm.Subscribe(s, taskId, models.TaskSubscribe, time.Now(), nil)
		if err != nil {
			c.log.Error().Stack().Err(err).Msg("")
		}
	}
}

func (c *SimpleCaster) Relay(msg models.WsResponseInfo) {
	if !c.enabled.Load() {
		return
	}

	c.addToQueue(msg)
}
