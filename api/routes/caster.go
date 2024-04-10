package routes

import (
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewCaster(recipientUsernames ...types.Username) types.BroadcasterAgent {
	recipients := util.Filter(util.MapToSlicePure(cmInstance.clientMap), func(c *Client) bool { return slices.Contains(recipientUsernames, c.user.GetUsername()) })

	newCaster := &unbufferedCaster{
		enabled:    false,
		recipients: recipients,
	}
	return newCaster
}

func (c *unbufferedCaster) Enable() {
	c.enabled = true
}

func (c unbufferedCaster) IsBuffered() bool {
	return true
}

func (c *unbufferedCaster) PushTaskUpdate(taskId types.TaskId, status string, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		MessageStatus: status,
		SubscribeKey:  subId(taskId),
		Content:       []wsM{{"result": result}},

		broadcastType: "task",
	}

	send(msg, c.recipients...)
}

func (c *unbufferedCaster) PushShareUpdate(username types.Username, newShareInfo types.Share) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		MessageStatus: "share_updated",
		SubscribeKey:  subId(username),
		Content:       []wsM{{"newShareInfo": newShareInfo}},
		Error:         "",
		broadcastType: "user",
	}

	send(msg, c.recipients...)
}

func (c unbufferedCaster) PushFileCreate(newFile types.WeblensFile) {
	if !c.enabled {
		return
	}
	fileInfo, err := newFile.FormatFileInfo(nil)
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

	fileInfo, err := updatedFile.FormatFileInfo(nil)
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
		MessageStatus: "file_moved",
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

func NewBufferedCaster() types.BufferedBroadcasterAgent {
	newCaster := &bufferedCaster{
		bufLimit:          100,
		buffer:            []wsResponse{},
		autoFlush:         false,
		autoFlushInterval: time.Second,
		bufLock:           &sync.Mutex{},
		enabled:           false,
	}

	return newCaster
}

func (c *bufferedCaster) AutoflushEnable() {
	c.enabled = true
	c.enableAutoflush()
}

func (c *bufferedCaster) Enable() {
	c.enabled = true
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
		MessageStatus: "file_created",
		SubscribeKey:  subId(newFile.GetParent().Id()),
		Content:       []wsM{{"fileInfo": fileInfo}},

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
		MessageStatus: "file_updated",
		SubscribeKey:  subId(updatedFile.Id()),
		Content:       []wsM{{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)

	if updatedFile.GetParent().Owner() == dataStore.WEBLENS_ROOT_USER {
		return
	}

	msg = wsResponse{
		MessageStatus: "file_updated",
		SubscribeKey:  subId(updatedFile.GetParent().Id()),
		Content:       []wsM{{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}

	// This is potentially problematic, if you are seeing issues with files not getting
	// websocket updates when they should START HERE

	// Immediately didn't work, I was right. Might revisit

	// var e bool
	// var em wsResponse
	// c.bufLock.Lock()
	// c.buffer, em, e = util.YoinkFunc(c.buffer, func(m wsResponse) bool {
	// 	return m.MessageStatus == "file_updated" && m.SubscribeKey == msg.SubscribeKey
	// })
	// c.bufLock.Unlock()

	// if e && len(em.Content) > 1 {
	// 	util.Error.Println("MIGHT BE BAD, REMOVED:", em)
	// }

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
		MessageStatus: "file_moved",
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
		MessageStatus: "file_deleted",
		SubscribeKey:  subId(deletedFile.GetParent().Id()),
		Content:       []wsM{{"fileId": deletedFile.Id()}},
		Error:         "",
		broadcastType: "folder",
	}

	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushTaskUpdate(taskId types.TaskId, status string, result types.TaskResult) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		MessageStatus: status,
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
		MessageStatus: "share_updated",
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

func (c *bufferedCaster) DisableAutoflush() {
	c.autoFlush = false
}

func (c *bufferedCaster) enableAutoflush() {
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
		return m.MessageStatus == msg.MessageStatus && m.SubscribeKey == msg.SubscribeKey
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

func send(r wsResponse, recipients ...*Client) {
	if len(recipients) != 0 {
		for _, c := range recipients {
			c.writeToClient(r)
		}
		return
	}

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
