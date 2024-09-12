package models

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/task"
)

var _ Broadcaster = (*SimpleCaster)(nil)
var _ Broadcaster = (*BufferedCaster)(nil)

type SimpleCaster struct {
	enabled bool
	cm      ClientManager
}

func (c *SimpleCaster) DisableAutoFlush() {
	// no-op
}

func (c *SimpleCaster) AutoFlushEnable() {
	// no-op
}

func (c *SimpleCaster) Flush() {
	// no-op
}

func (c *SimpleCaster) Close() {
	// c.enabled = false
}

type BufferedCaster struct {
	bufLimit          int
	buffer            []WsResponseInfo
	autoFlush         atomic.Bool
	enabled           atomic.Bool
	autoFlushInterval time.Duration
	bufLock           sync.Mutex

	cm ClientManager
}

func NewSimpleCaster(cm ClientManager) *SimpleCaster {
	newCaster := &SimpleCaster{
		enabled: true,
		cm:      cm,
	}
	return newCaster
}

func (c *SimpleCaster) Enable() {
	c.enabled = true
}

func (c *SimpleCaster) Disable() {
	c.enabled = false
}

func (c *SimpleCaster) IsBuffered() bool {
	return false
}

func (c *SimpleCaster) IsEnabled() bool {
	return c.enabled
}

func (c *SimpleCaster) PushWeblensEvent(eventTag string, content ...WsC) {
	msg := WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  "WEBLENS",
		BroadcastType: ServerEvent,
	}

	if len(content) != 0 {
		msg.Content = content[0]
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.TaskId(),
		Content:       WsC(result),
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) PushPoolUpdate(
	pool task.Pool, event string, result task.TaskResult,
) {
	if !c.enabled {
		return
	}

	if pool.IsGlobal() {
		log.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  pool.ID(),
		Content:       WsC(result),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	c.cm.Send(msg)
	// c.c.cm.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (c *SimpleCaster) PushShareUpdate(username Username, newShareInfo Share) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "share_updated",
		SubscribeKey:  username,
		Content:       WsC{"newShareInfo": newShareInfo},
		BroadcastType: UserSubscribe,
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) PushFileCreate(newFile *fileTree.WeblensFileImpl) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:     "file_created",
		SubscribeKey: newFile.GetParentId(),
		Content:      WsC{"fileInfo": newFile},

		BroadcastType: FolderSubscribe,
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  updatedFile.ID(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
	}

	c.cm.Send(msg)

	if updatedFile.GetParent() == nil || updatedFile.GetParent().ID() == "ROOT" {
		return
	}

	msg = WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  updatedFile.GetParentId(),
		Content:       WsC{"fileInfo": updatedFile},
		BroadcastType: FolderSubscribe,
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  preMoveFile.GetParentId(),
		Content:       WsC{"oldId": preMoveFile.ID(), "newFile": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
	}
	c.cm.Send(msg)

	msg = WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  postMoveFile.GetParentId(),
		Content:       WsC{"oldId": preMoveFile.ID(), "newFile": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
	}
	c.cm.Send(msg)
}

func (c *SimpleCaster) PushFileDelete(deletedFile *fileTree.WeblensFileImpl) {
	if !c.enabled {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_deleted",
		SubscribeKey:  deletedFile.GetParent().ID(),
		Content:       WsC{"fileId": deletedFile.ID()},
		BroadcastType: FolderSubscribe,
	}

	c.cm.Send(msg)
}

func (c *SimpleCaster) FolderSubToTask(folder fileTree.FileId, taskId task.Id) {
	if !c.enabled {
		return
	}

	subs := c.cm.GetSubscribers(FolderSubscribe, folder)

	for _, s := range subs {
		_, _, err := c.cm.Subscribe(s, taskId, TaskSubscribe, time.Now(), nil)
		if err != nil {
			log.ShowErr(err)
		}
	}
}

// func (c *SimpleCaster) UnsubTask(task *task.Task) {
// 	if !c.enabled {
// 		return
// 	}
//
// 	subs := c.cm.GetSubscribers(FolderSubscribe, SubId(task.TaskId()))
// 	for _, s := range subs {
// 		s.unsubscribe(SubId(task.TaskId()))
// 	}
// }

func (c *SimpleCaster) Relay(msg WsResponseInfo) {
	if !c.enabled {
		return
	}

	c.cm.Send(msg)
}

// NewBufferedCaster Gets a new buffered caster with the auto-flusher pre-enabled.
// c.Close() must be called when this caster is no longer in use to
// release the flusher
func NewBufferedCaster(cm ClientManager) *BufferedCaster {
	// local := InstanceService.GetLocal()
	// if local == nil || local.ServerRole() != weblens.CoreServer {
	// 	return &bufferedCaster{enabled: atomic.Bool{}, autoFlushInterval: time.Hour}
	// }
	newCaster := &BufferedCaster{
		bufLimit:          100,
		buffer:            []WsResponseInfo{},
		autoFlushInterval: time.Second,
		cm:                cm,
	}

	newCaster.enabled.Store(true)
	newCaster.enableAutoFlush()

	return newCaster
}

func (c *BufferedCaster) AutoFlushEnable() {
	c.enabled.Store(true)
	c.enableAutoFlush()
}

func (c *BufferedCaster) Enable() {
	c.enabled.Store(true)
}

func (c *BufferedCaster) Disable() {
	c.enabled.Store(false)
}

func (c *BufferedCaster) Close() {
	if !c.enabled.Load() {
		log.ErrTrace(werror.ErrCasterDoubleClose)
		return
	}

	c.Flush()
	c.autoFlush.Store(false)
	c.enabled.Store(false)
}

func (c *BufferedCaster) IsBuffered() bool {
	return true
}

func (c *BufferedCaster) IsEnabled() bool {
	return c.enabled.Load()
}

func (c *BufferedCaster) PushWeblensEvent(eventTag string, content ...WsC) {
	msg := WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  "WEBLENS",
		BroadcastType: ServerEvent,
	}

	if len(content) != 0 {
		msg.Content = content[0]
	}

	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushFileCreate(newFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:     "file_created",
		SubscribeKey: newFile.GetParentId(),
		Content:      WsC{"fileInfo": newFile},

		BroadcastType: FolderSubscribe,
	}
	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  updatedFile.ID(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
	}

	c.bufferAndFlush(msg)

	if updatedFile.GetParent() == nil || updatedFile.GetParent().ID() == "ROOT" {
		return
	}

	msg = WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  updatedFile.GetParentId(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  preMoveFile.GetParentId(),
		Content:       WsC{"oldId": preMoveFile.ID(), "newFile": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
	}
	c.bufferAndFlush(msg)

	msg = WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  postMoveFile.GetParentId(),
		Content:       WsC{"oldId": preMoveFile.ID(), "newFile": postMoveFile},
		Error:         "",
		BroadcastType: FolderSubscribe,
	}
	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushFileDelete(deletedFile *fileTree.WeblensFileImpl) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "file_deleted",
		SubscribeKey:  deletedFile.GetParent().ID(),
		Content:       WsC{"fileId": deletedFile.ID()},
		BroadcastType: FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.TaskId(),
		Content:       WsC(result),
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
	}

	c.bufferAndFlush(msg)

	msg = WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.GetTaskPool().GetRootPool().ID(),
		Content:       result,
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
	}

	c.bufferAndFlush(msg)

}

func (c *BufferedCaster) PushPoolUpdate(
	pool task.Pool, event string, result task.TaskResult,
) {
	if pool.IsGlobal() {
		log.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  pool.ID(),
		Content:       WsC(result),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) PushShareUpdate(username Username, newShareInfo Share) {
	if !c.enabled.Load() {
		return
	}

	msg := WsResponseInfo{
		EventTag:      "share_updated",
		SubscribeKey:  username,
		Content:       WsC{"newShareInfo": newShareInfo},
		BroadcastType: "user",
	}

	c.bufferAndFlush(msg)
}

func (c *BufferedCaster) Flush() {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	if !c.enabled.Load() || len(c.buffer) == 0 {
		return
	}

	for _, r := range c.buffer {
		c.cm.Send(r)
	}

	c.buffer = []WsResponseInfo{}
}

func (c *BufferedCaster) DisableAutoFlush() {
	c.autoFlush.Store(false)
}

// FolderSubToTask Subscribes any subscribers of a folder to a task (presumably one that pertains to that folder)
func (c *BufferedCaster) FolderSubToTask(folder fileTree.FileId, taskId task.Id) {
	subs := c.cm.GetSubscribers(FolderSubscribe, folder)

	for _, s := range subs {
		_, _, err := c.cm.Subscribe(s, taskId, TaskSubscribe, time.Now(), nil)
		if err != nil {
			log.ShowErr(err)
		}
	}
}

// func (c *BufferedCaster) UnsubTask(task *task.Task) {
// 	subs := c.cm.GetSubscribers(FolderSubscribe, SubId(task.TaskId()))
// 	for _, s := range subs {
// 		s.unsubscribe(SubId(task.TaskId()))
// 	}
// }

func (c *BufferedCaster) Relay(msg WsResponseInfo) {
	if !c.enabled.Load() {
		return
	}

	c.cm.Send(msg)
}

func (c *BufferedCaster) enableAutoFlush() {
	c.autoFlush.Store(true)
	go c.flusher()
}

func (c *BufferedCaster) flusher() {
	for c.autoFlush.Load() {
		time.Sleep(c.autoFlushInterval)
		c.Flush()
	}
}

func (c *BufferedCaster) bufferAndFlush(msg WsResponseInfo) {
	c.bufLock.Lock()
	c.buffer = append(c.buffer, msg)

	if c.autoFlush.Load() && len(c.buffer) >= c.bufLimit {
		c.bufLock.Unlock()
		c.Flush()
		return
	}
	c.bufLock.Unlock()
}

type BasicCaster interface {
	PushWeblensEvent(eventTag string, content ...WsC)

	PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media)
	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
}

type Broadcaster interface {
	BasicCaster
	PushFileCreate(newFile *fileTree.WeblensFileImpl)
	PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl)
	PushFileDelete(deletedFile *fileTree.WeblensFileImpl)
	PushShareUpdate(username Username, newShareInfo Share)
	Enable()
	Disable()
	IsEnabled() bool
	IsBuffered() bool

	FolderSubToTask(folder fileTree.FileId, taskId task.Id)
	// UnsubTask(task *task.Task)
	DisableAutoFlush()
	AutoFlushEnable()
	Flush()

	Relay(msg WsResponseInfo)

	// Close flush, release the auto-flusher, and disable the caster
	Close()
}

// WsC is the generic WebSocket Content container
type WsC map[string]any
type SubId = string
type WsAction string
type ClientType string

const (
	// UserSubscribe does not actually get "subscribed" to, it is automatically tracked for every websocket
	// connection made, and only sends updates to that specific user when needed
	UserSubscribe WsAction = "user_subscribe"

	FolderSubscribe   WsAction = "folder_subscribe"
	ServerEvent       WsAction = "server_event"
	TaskSubscribe     WsAction = "task_subscribe"
	PoolSubscribe     WsAction = "pool_subscribe"
	TaskTypeSubscribe WsAction = "task_type_subscribe"
	Unsubscribe       WsAction = "unsubscribe"
	ScanDirectory     WsAction = "scan_directory"
	CancelTask        WsAction = "cancel_task"
	ReportError       WsAction = "show_web_error"
)

const (
	WebClient    ClientType = "webClient"
	RemoteClient ClientType = "remoteClient"
)

type Subscription struct {
	Type WsAction
	Key  SubId
	When time.Time
}

type WsResponseInfo struct {
	EventTag      string   `json:"eventTag"`
	SubscribeKey  SubId    `json:"subscribeKey"`
	TaskType      string   `json:"taskType,omitempty"`
	Content       WsC      `json:"content"`
	Error         string   `json:"error,omitempty"`
	BroadcastType WsAction `json:"broadcastType,omitempty"`
}

type WsRequestInfo struct {
	Action  WsAction `json:"action"`
	SentAt  int64    `json:"sentAt"`
	Content string   `json:"content"`
}

// WsR WebSocket Request interface
type WsR interface {
	GetKey() SubId
	Action() WsAction
}

const (
	StartupProgressEvent = "startup_progress"
	TaskCreatedEvent     = "task_created"
	TaskCompleteEvent    = "task_complete"
	SubTaskCompleteEvent = "sub_task_complete"
	TaskFailedEvent      = "task_failure"
	TaskCanceledEvent    = "task_canceled"
	PoolCreatedEvent     = "pool_created"
	PoolCompleteEvent    = "pool_complete"
	PoolCancelledEvent   = "pool_cancelled"
	ScanCompleteEvent    = "scan_complete"
	ZipProgressEvent     = "create_zip_progress"
	ZipCompleteEvent     = "zip_complete"
)
