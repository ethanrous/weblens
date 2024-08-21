package websocket

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/http"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

type unbufferedCaster struct {
	enabled bool
}

type bufferedCaster struct {
	bufLimit          int
	buffer            []types.WsResponseInfo
	autoFlush         atomic.Bool
	enabled           atomic.Bool
	autoFlushInterval time.Duration
	bufLock sync.Mutex
}

func NewCaster() BroadcasterAgent {
	serverInfo := InstanceService.GetLocal()
	if serverInfo == nil {
		return &unbufferedCaster{enabled: false}
	}

	newCaster := &unbufferedCaster{
		enabled: true,
	}
	return newCaster
}

func (c *unbufferedCaster) Enable() {
	c.enabled = true
}

func (c *unbufferedCaster) Disable() {
	c.enabled = false
}

func (c *unbufferedCaster) IsBuffered() bool {
	return false
}

func (c *unbufferedCaster) IsEnabled() bool {
	return c.enabled
}

func (c *unbufferedCaster) PushWeblensEvent(eventTag string) {
	msg := types.WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  types.SubId("WEBLENS"),
		BroadcastType: types.ServerEvent,
	}

	send(msg)
}

func (c *unbufferedCaster) PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.TaskId()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	if pool.IsGlobal() {
		wlog.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(pool.ID()),
		Content:       types.WsC(result),
		TaskType:      pool.CreatedInTask().TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (c *unbufferedCaster) PushShareUpdate(username types.Username, newShareInfo weblens.Share) {
	if !c.enabled {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "share_updated",
		SubscribeKey:  types.SubId(username),
		Content:       types.WsC{"newShareInfo": newShareInfo},
		BroadcastType: types.UserSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileCreate(newFile *fileTree.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(http.UserService.Get("WEBLENS")).SetRequestMode(
		dataStore.WebsocketFileUpdate,
	)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_created",
		SubscribeKey:  types.SubId(newFile.GetParent().ID()),
		Content:       types.WsC{"fileInfo": fileInfo},
		BroadcastType: types.FolderSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileUpdate(updatedFile *fileTree.WeblensFile) {
	if !c.enabled {
		return
	}

	if updatedFile.Owner() == http.UserService.Get("WEBLENS") {
		return
	}

	acc := dataStore.NewAccessMeta(http.UserService.Get("WEBLENS")).SetRequestMode(
		dataStore.WebsocketFileUpdate,
	)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  types.SubId(updatedFile.ID()),
		Content:       types.WsC{"fileInfo": fileInfo},
		BroadcastType: types.FolderSubscribe,
	}

	send(msg)

	if updatedFile.GetParent().Owner() == http.UserService.Get("WEBLENS") {
		return
	}

	msg = types.WsResponseInfo{
		EventTag:     "file_updated",
		SubscribeKey: types.SubId(updatedFile.GetParent().ID()),
		Content:      types.WsC{"fileInfo": fileInfo},
		BroadcastType: types.FolderSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileMove(preMoveFile *fileTree.WeblensFile, postMoveFile *fileTree.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(http.UserService.Get("WEBLENS")).SetRequestMode(
		dataStore.WebsocketFileUpdate,
	)

	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  types.SubId(preMoveFile.GetParent().ID()),
		Content:       types.WsC{"oldId": preMoveFile.ID(), "newFile": postInfo},
		BroadcastType: types.FolderSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileDelete(deletedFile *fileTree.WeblensFile) {
	if !c.enabled {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_deleted",
		SubscribeKey:  types.SubId(deletedFile.GetParent().ID()),
		Content:       types.WsC{"fileId": deletedFile.ID()},
		BroadcastType: types.FolderSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) FolderSubToTask(folder types.FileId, taskId types.TaskId) {
	if !c.enabled {
		return
	}

	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))

	for _, s := range subs {
		s.Subscribe(types.SubId(taskId), types.TaskSubscribe, nil)
	}
}

func (c *unbufferedCaster) FolderSubToPool(folder types.FileId, poolId types.TaskId) {
	if !c.enabled {
		return
	}

	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))

	for _, s := range subs {
		s.Subscribe(types.SubId(poolId), types.PoolSubscribe, nil)
	}
}

func (c *unbufferedCaster) UnsubTask(task types.Task) {
	if !c.enabled {
		return
	}

	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(types.SubId(task.TaskId()))
	}
}

func (c *unbufferedCaster) Relay(msg types.WsResponseInfo) {
	if !c.enabled {
		return
	}

	send(msg)
}

// NewBufferedCaster Gets a new buffered caster with the auto-flusher pre-enabled.
// c.Close() must be called when this caster is no longer in use to
// release the flusher
func NewBufferedCaster() BufferedBroadcasterAgent {
	local := InstanceService.GetLocal()
	if local == nil || local.ServerRole() != Core {
		return &bufferedCaster{enabled: atomic.Bool{}, autoFlushInterval: time.Hour}
	}
	newCaster := &bufferedCaster{
		enabled:           atomic.Bool{},
		bufLimit:          100,
		buffer:            []types.WsResponseInfo{},
		autoFlushInterval: time.Second,
		bufLock: sync.Mutex{},
	}

	newCaster.enabled.Store(true)
	newCaster.enableAutoFlush()

	return newCaster
}

func (c *bufferedCaster) AutoFlushEnable() {
	c.enabled.Store(true)
	c.enableAutoFlush()
}

func (c *bufferedCaster) Enable() {
	c.enabled.Store(true)
}

func (c *bufferedCaster) Disable() {
	c.enabled.Store(false)
}

func (c *bufferedCaster) Close() {
	if !c.enabled.Load() {
		wlog.ErrTrace(werror.ErrCasterDoubleClose)
		return
	}

	c.Flush()
	c.autoFlush.Store(false)
	c.enabled.Store(false)
}

func (c *bufferedCaster) IsBuffered() bool {
	return true
}

func (c *bufferedCaster) IsEnabled() bool {
	return c.enabled.Load()
}

func (c *bufferedCaster) PushWeblensEvent(eventTag string) {
	msg := types.WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  types.SubId("WEBLENS"),
		BroadcastType: types.ServerEvent,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileCreate(newFile *fileTree.WeblensFile) {
	if !c.enabled.Load() {
		return
	}

	if newFile.Owner() == http.UserService.Get("WEBLENS") {
		return
	}

	acc := dataStore.NewAccessMeta(nil).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:     "file_created",
		SubscribeKey: types.SubId(newFile.GetParent().ID()),
		Content:      types.WsC{"fileInfo": fileInfo},

		BroadcastType: types.FolderSubscribe,
	}
	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileUpdate(updatedFile *fileTree.WeblensFile) {
	if !c.enabled.Load() {
		return
	}

	if updatedFile.Owner() == http.UserService.Get("WEBLENS") {
		return
	}

	acc := dataStore.NewAccessMeta(http.UserService.Get("WEBLENS")).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:     "file_updated",
		SubscribeKey: types.SubId(updatedFile.ID()),
		Content:      types.WsC{"fileInfo": fileInfo},

		BroadcastType: types.FolderSubscribe,
	}
	c.bufferAndFlush(msg)

	if updatedFile.GetParent().Owner().GetUsername() == "WEBLENS" {
		return
	}

	msg = types.WsResponseInfo{
		EventTag:     "file_updated",
		SubscribeKey: types.SubId(updatedFile.GetParent().ID()),
		Content:      types.WsC{"fileInfo": fileInfo},

		BroadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileMove(preMoveFile *fileTree.WeblensFile, postMoveFile *fileTree.WeblensFile) {
	if !c.enabled.Load() {
		return
	}

	acc := dataStore.NewAccessMeta(http.UserService.Get("WEBLENS")).SetRequestMode(dataStore.WebsocketFileUpdate)
	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_moved",
		SubscribeKey:  types.SubId(preMoveFile.GetParent().ID()),
		Content:       types.WsC{"oldId": preMoveFile.ID(), "newFile": postInfo},
		Error:         "",
		BroadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileDelete(deletedFile *fileTree.WeblensFile) {
	if !c.enabled.Load() {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "file_deleted",
		SubscribeKey:  types.SubId(deletedFile.GetParent().ID()),
		Content:       types.WsC{"fileId": deletedFile.ID()},
		BroadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled.Load() {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.TaskId()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	c.bufferAndFlush(msg)

	msg = types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.GetTaskPool().GetRootPool().ID()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	c.bufferAndFlush(msg)

}

func (c *bufferedCaster) PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult) {
	if pool.IsGlobal() {
		wlog.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(pool.ID()),
		Content:       types.WsC(result),
		TaskType:      pool.CreatedInTask().TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushShareUpdate(username types.Username, newShareInfo weblens.Share) {
	if !c.enabled.Load() {
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      "share_updated",
		SubscribeKey:  types.SubId(username),
		Content:       types.WsC{"newShareInfo": newShareInfo},
		BroadcastType: "user",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) Flush() {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	if !c.enabled.Load() || len(c.buffer) == 0 {
		return
	}

	for _, r := range c.buffer {
		send(r)
	}

	c.buffer = []types.WsResponseInfo{}
}

func (c *bufferedCaster) DisableAutoFlush() {
	c.autoFlush.Store(false)
}

// FolderSubToTask Subscribes any subscribers of a folder to a task (presumably one that pertains to that folder)
func (c *bufferedCaster) FolderSubToTask(folder types.FileId, taskId types.TaskId) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))

	for _, s := range subs {
		s.Subscribe(types.SubId(taskId), types.TaskSubscribe, nil)
	}
}

func (c *bufferedCaster) FolderSubToPool(folder types.FileId, poolId types.TaskId) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))

	for _, s := range subs {
		s.Subscribe(types.SubId(poolId), types.PoolSubscribe, nil)
	}
}

func (c *bufferedCaster) UnsubTask(task types.Task) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(types.SubId(task.TaskId()))
	}
}

func (c *bufferedCaster) enableAutoFlush() {
	c.autoFlush.Store(true)
	go c.flusher()
}

func (c *bufferedCaster) flusher() {
	for c.autoFlush.Load() {
		time.Sleep(c.autoFlushInterval)
		c.Flush()
	}
}

func (c *bufferedCaster) bufferAndFlush(msg types.WsResponseInfo) {
	c.bufLock.Lock()
	c.buffer = append(c.buffer, msg)

	if c.autoFlush.Load() && len(c.buffer) >= c.bufLimit {
		c.bufLock.Unlock()
		c.Flush()
		return
	}
	c.bufLock.Unlock()
}

func send(r types.WsResponseInfo) {
	defer internal.RecoverPanic("Panic caught while broadcasting")

	if r.SubscribeKey == "" {
		wlog.Error.Println("Trying to broadcast on empty key")
		return
	}

	var clients []types.Client
	if !InstanceService.IsLocalLoaded() || r.BroadcastType == types.ServerEvent {
		clients = types.SERV.ClientManager.GetAllClients()
	} else {
		clients = types.SERV.ClientManager.GetSubscribers(r.BroadcastType, r.SubscribeKey)
		clients = internal.OnlyUnique(clients)
	}

	if r.BroadcastType == types.TaskSubscribe {
		clients = append(
			clients, types.SERV.ClientManager.GetSubscribers(
				types.TaskTypeSubscribe,
				types.SubId(r.TaskType),
			)...,
		)
	}

	if len(clients) != 0 {
		for _, c := range clients {
			c.(*WsClient).Send(r)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", r.SubscribeKey)
		return
	}
}

type BasicCaster interface {
	PushWeblensEvent(eventTag string)

	PushFileUpdate(updatedFile *fileTree.WeblensFile)
	PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult)
	PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult)
}

type BroadcasterAgent interface {
	BasicCaster
	PushFileCreate(newFile *fileTree.WeblensFile)
	PushFileMove(preMoveFile *fileTree.WeblensFile, postMoveFile *fileTree.WeblensFile)
	PushFileDelete(deletedFile *fileTree.WeblensFile)
	PushShareUpdate(username types.Username, newShareInfo weblens.Share)
	Enable()
	Disable()
	IsEnabled() bool
	IsBuffered() bool

	FolderSubToTask(folder types.FileId, taskId types.TaskId)
	FolderSubToPool(folder types.FileId, poolId types.TaskId)
	UnsubTask(task types.Task)
}

type BufferedBroadcasterAgent interface {
	BroadcasterAgent
	DisableAutoFlush()
	AutoFlushEnable()
	Flush()

	// Close flush, release the auto-flusher, and disable the caster
	Close()
}

