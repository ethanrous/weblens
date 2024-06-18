package routes

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewCaster() types.BroadcasterAgent {
	serverInfo := types.SERV.InstanceService.GetLocal()
	if serverInfo == nil || serverInfo.ServerRole() != types.Core {
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

func (c *unbufferedCaster) IsBuffered() bool {
	return false
}

func (c *unbufferedCaster) PushTaskUpdate(taskId types.TaskId, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:     string(event),
		SubscribeKey: types.SubId(taskId),
		Content:      []types.WsMsg{{"result": result}},

		broadcastType: types.TaskSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushShareUpdate(username types.Username, newShareInfo types.Share) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "share_updated",
		SubscribeKey:  types.SubId(username),
		Content:       []types.WsMsg{{"newShareInfo": newShareInfo}},
		Error:         "",
		broadcastType: "user",
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileCreate(newFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser, newFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	types.SERV.ClientManager.Broadcast(types.FolderSubscribe, types.SubId(newFile.GetParent().ID()), "file_created", []types.WsMsg{{"fileInfo": fileInfo}})
}

func (c *unbufferedCaster) PushFileUpdate(updatedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	if updatedFile.Owner() == dataStore.WeblensRootUser {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser, updatedFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	types.SERV.ClientManager.Broadcast(types.FolderSubscribe, types.SubId(updatedFile.ID()), "file_updated", []types.WsMsg{{"fileInfo": fileInfo}})

	if updatedFile.GetParent().Owner() == dataStore.WeblensRootUser {
		return
	}
	types.SERV.ClientManager.Broadcast(types.FolderSubscribe, types.SubId(updatedFile.GetParent().ID()), "file_updated", []types.WsMsg{{"fileInfo": fileInfo}})
}

func (c *unbufferedCaster) PushFileMove(preMoveFile types.WeblensFile, postMoveFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser, postMoveFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)

	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:      "file_moved",
		SubscribeKey:  types.SubId(preMoveFile.GetParent().ID()),
		Content:       []types.WsMsg{{"oldId": preMoveFile.ID(), "newFile": postInfo}},
		Error:         "",
		broadcastType: types.FolderSubscribe,
	}

	send(msg)
}

func (c *unbufferedCaster) PushFileDelete(deletedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	content := []types.WsMsg{{"fileId": deletedFile.ID()}}
	types.SERV.ClientManager.Broadcast(types.FolderSubscribe, types.SubId(deletedFile.GetParent().ID()), "file_deleted", content)
}

func (c *unbufferedCaster) FolderSubToTask(folder types.FileId, taskId types.TaskId) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))
	for _, s := range subs {
		s.Subscribe(types.SubId(taskId), types.TaskSubscribe, nil, types.SERV.FileTree)
	}
}

func (c *unbufferedCaster) UnsubTask(task types.Task) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(types.SubId(task.TaskId()))
	}
}

// Get a new buffered caster with the auto-flusher pre-enabled.
// c.Close() must be called when this caster is no longer in use to
// release the flusher
func NewBufferedCaster() types.BufferedBroadcasterAgent {
	local := types.SERV.InstanceService.GetLocal()
	if local == nil || local.ServerRole() != types.Core {
		return &bufferedCaster{enabled: false, autoFlushInterval: time.Hour}
	}
	newCaster := &bufferedCaster{
		enabled:           true,
		bufLimit:          100,
		buffer:            []wsResponse{},
		autoFlushInterval: time.Second,
		bufLock:           &sync.Mutex{},
	}

	newCaster.enableAutoFlush()

	return newCaster
}

func (c *bufferedCaster) AutoFlushEnable() {
	c.enabled = true
	c.enableAutoFlush()
}

func (c *bufferedCaster) Enable() {
	c.enabled = true
}

func (c *bufferedCaster) Close() {
	if !c.enabled {
		util.ErrTrace(ErrCasterDoubleClose)
		return
	}

	c.Flush()
	c.autoFlush = false
	c.enabled = false
}

func (c *bufferedCaster) IsBuffered() bool {
	return true
}

func (c *bufferedCaster) PushFileCreate(newFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	if newFile.Owner() == dataStore.WeblensRootUser {
		return
	}

	acc := dataStore.NewAccessMeta(nil, newFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:     "file_created",
		SubscribeKey: types.SubId(newFile.GetParent().ID()),
		Content:      []types.WsMsg{{"fileInfo": fileInfo}},

		broadcastType: types.FolderSubscribe,
	}
	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileUpdate(updatedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	if updatedFile.Owner() == dataStore.WeblensRootUser {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser, updatedFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:     "file_updated",
		SubscribeKey: types.SubId(updatedFile.ID()),
		Content:      []types.WsMsg{{"fileInfo": fileInfo}},

		broadcastType: types.FolderSubscribe,
	}
	c.bufferAndFlush(msg)

	if updatedFile.GetParent().Owner() == dataStore.WeblensRootUser {
		return
	}

	msg = wsResponse{
		EventTag:     "file_updated",
		SubscribeKey: types.SubId(updatedFile.GetParent().ID()),
		Content:      []types.WsMsg{{"fileInfo": fileInfo}},

		broadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileMove(preMoveFile types.WeblensFile, postMoveFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser, postMoveFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:      "file_moved",
		SubscribeKey:  types.SubId(preMoveFile.GetParent().ID()),
		Content:       []types.WsMsg{{"oldId": preMoveFile.ID(), "newFile": postInfo}},
		Error:         "",
		broadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileDelete(deletedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "file_deleted",
		SubscribeKey:  types.SubId(deletedFile.GetParent().ID()),
		Content:       []types.WsMsg{{"fileId": deletedFile.ID()}},
		Error:         "",
		broadcastType: types.FolderSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushTaskUpdate(taskId types.TaskId, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(taskId),
		Content:       []types.WsMsg{types.WsMsg(result)},
		Error:         "",
		broadcastType: types.TaskSubscribe,
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushShareUpdate(username types.Username, newShareInfo types.Share) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "share_updated",
		SubscribeKey:  types.SubId(username),
		Content:       []types.WsMsg{{"newShareInfo": newShareInfo}},
		Error:         "",
		broadcastType: "user",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) Flush() {
	if !c.enabled || len(c.buffer) == 0 {
		return
	}

	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	for _, r := range c.buffer {
		send(r)
	}

	c.buffer = []wsResponse{}
}

func (c *bufferedCaster) DropBuffer() {
	c.buffer = []wsResponse{}
}

func (c *bufferedCaster) DisableAutoFlush() {
	c.autoFlush = false
}

// Subscribe any subscribers of a folder to a task (presumably one that pertains to that folder)
func (c *bufferedCaster) FolderSubToTask(folder types.FileId, task types.TaskId) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(folder))
	for _, s := range subs {
		s.Subscribe(types.SubId(task), types.TaskSubscribe, nil, types.SERV.FileTree)
	}
}

func (c *bufferedCaster) UnsubTask(task types.Task) {
	subs := types.SERV.ClientManager.GetSubscribers(types.FolderSubscribe, types.SubId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(types.SubId(task.TaskId()))
	}
}

func (c *bufferedCaster) enableAutoFlush() {
	c.autoFlush = true
	go c.flusher()
}

func (c *bufferedCaster) flusher() {
	for c.autoFlush {
		time.Sleep(c.autoFlushInterval)
		c.Flush()
	}
}

func (c *bufferedCaster) bufferAndFlush(msg wsResponse) {
	c.bufLock.Lock()
	index := util.Find(c.buffer, func(m wsResponse) bool {
		return m.EventTag == msg.EventTag && m.SubscribeKey == msg.SubscribeKey
	})
	if index == -1 {
		c.buffer = append(c.buffer, msg)
	} else {
		c.buffer[index].Content = append(c.buffer[index].Content, msg.Content...)
	}
	shouldFlush := c.autoFlush && len(c.buffer) >= c.bufLimit
	c.bufLock.Unlock()

	if shouldFlush {
		c.Flush()
	}
}

func send(r wsResponse) {
	if r.SubscribeKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}
	defer util.RecoverPanic("Panic caught while broadcasting: %v")

	clients := types.SERV.ClientManager.GetSubscribers(r.broadcastType, types.SubId(r.SubscribeKey))
	clients = util.OnlyUnique(clients)

	if len(clients) != 0 {
		for _, c := range clients {
			c.(*client).writeToClient(r)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", r.SubscribeKey)
		return
	}
}
