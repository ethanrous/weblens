package routes

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewCaster() types.BroadcasterAgent {
	serverInfo := dataStore.GetServerInfo()
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

func (c unbufferedCaster) IsBuffered() bool {
	return false
}

func (c *unbufferedCaster) PushTaskUpdate(taskId types.TaskId, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:     string(event),
		SubscribeKey: subId(taskId),
		Content:      []wsM{{"result": result}},

		broadcastType: "task",
	}

	send(msg)
}

func (c *unbufferedCaster) PushShareUpdate(username types.Username, newShareInfo types.Share) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "share_updated",
		SubscribeKey:  subId(username),
		Content:       []wsM{{"newShareInfo": newShareInfo}},
		Error:         "",
		broadcastType: "user",
	}

	send(msg)
}

func (c unbufferedCaster) PushFileCreate(newFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	cmInstance.Broadcast("folder", subId(newFile.GetParent().Id()), "file_created", []wsM{{"fileInfo": fileInfo}})
}

func (c unbufferedCaster) PushFileUpdate(updatedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	if updatedFile.Owner() == dataStore.WEBLENS_ROOT_USER {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	cmInstance.Broadcast("folder", subId(updatedFile.Id()), "file_updated", []wsM{{"fileInfo": fileInfo}})

	if updatedFile.GetParent().Owner() == dataStore.WEBLENS_ROOT_USER {
		return
	}
	cmInstance.Broadcast("folder", subId(updatedFile.GetParent().Id()), "file_updated", []wsM{{"fileInfo": fileInfo}})
}

func (c unbufferedCaster) PushFileMove(preMoveFile types.WeblensFile, postMoveFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER).SetRequestMode(dataStore.WebsocketFileUpdate)

	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:      "file_moved",
		SubscribeKey:  subId(preMoveFile.GetParent().Id()),
		Content:       []wsM{{"oldId": preMoveFile.Id(), "newFile": postInfo}},
		Error:         "",
		broadcastType: "folder",
	}

	send(msg)
}

func (c unbufferedCaster) PushFileDelete(deletedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	content := []wsM{{"fileId": deletedFile.Id()}}
	cmInstance.Broadcast("folder", subId(deletedFile.GetParent().Id()), "file_deleted", content)
}

func (c unbufferedCaster) FolderSubToTask(folder types.FileId, task types.TaskId) {
	subs := cmInstance.GetSubscribers(SubFolder, subId(folder))
	for _, s := range subs {
		s.Subscribe(SubTask, subId(task), nil)
	}
}

func (c unbufferedCaster) UnsubTask(task types.Task) {
	subs := cmInstance.GetSubscribers(SubFolder, subId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(subId(task.TaskId()))
	}
}

// Get a new buffered caster with the auto-flusher pre-enabled.
// c.Close() must be called when this caster is no longer in use to
// release the flusher
func NewBufferedCaster() types.BufferedBroadcasterAgent {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil || srvInfo.ServerRole() != types.Core {
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

func (c bufferedCaster) IsBuffered() bool {
	return true
}

func (c *bufferedCaster) PushFileCreate(newFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(nil).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := newFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:     "file_created",
		SubscribeKey: subId(newFile.GetParent().Id()),
		Content:      []wsM{{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileUpdate(updatedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	if updatedFile.Owner() == dataStore.WEBLENS_ROOT_USER {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:     "file_updated",
		SubscribeKey: subId(updatedFile.Id()),
		Content:      []wsM{{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)

	if updatedFile.GetParent().Owner() == dataStore.WEBLENS_ROOT_USER {
		return
	}

	msg = wsResponse{
		EventTag:     "file_updated",
		SubscribeKey: subId(updatedFile.GetParent().Id()),
		Content:      []wsM{{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileMove(preMoveFile types.WeblensFile, postMoveFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	acc := dataStore.NewAccessMeta(dataStore.WEBLENS_ROOT_USER).SetRequestMode(dataStore.WebsocketFileUpdate)
	postInfo, err := postMoveFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	msg := wsResponse{
		EventTag:      "file_moved",
		SubscribeKey:  subId(preMoveFile.GetParent().Id()),
		Content:       []wsM{{"oldId": preMoveFile.Id(), "newFile": postInfo}},
		Error:         "",
		broadcastType: "folder",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileDelete(deletedFile types.WeblensFile) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "file_deleted",
		SubscribeKey:  subId(deletedFile.GetParent().Id()),
		Content:       []wsM{{"fileId": deletedFile.Id()}},
		Error:         "",
		broadcastType: "folder",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushTaskUpdate(taskId types.TaskId, event types.TaskEvent, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      string(event),
		SubscribeKey:  subId(taskId),
		Content:       []wsM{wsM(result)},
		Error:         "",
		broadcastType: "task",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushShareUpdate(username types.Username, newShareInfo types.Share) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		EventTag:      "share_updated",
		SubscribeKey:  subId(username),
		Content:       []wsM{{"newShareInfo": newShareInfo}},
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
func (c bufferedCaster) FolderSubToTask(folder types.FileId, task types.TaskId) {
	subs := cmInstance.GetSubscribers(SubFolder, subId(folder))
	for _, s := range subs {
		s.Subscribe(SubTask, subId(task), nil)
	}
}

func (c bufferedCaster) UnsubTask(task types.Task) {
	subs := cmInstance.GetSubscribers(SubFolder, subId(task.TaskId()))
	for _, s := range subs {
		s.Unsubscribe(subId(task.TaskId()))
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

	clients := cmInstance.GetSubscribers(subType(r.broadcastType), subId(r.SubscribeKey))
	clients = util.OnlyUnique(clients)

	if len(clients) != 0 {
		for _, c := range clients {
			c.writeToClient(r)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", r.SubscribeKey)
		return
	}
}
