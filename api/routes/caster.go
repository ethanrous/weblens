package routes

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func NewBufferedCaster() BufferedBroadcasterAgent {
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

func (c *bufferedCaster) Enable(autoFlush bool) {
	c.enabled = true
	if autoFlush {
		c.enableAutoflush()
	}
}

func (c *bufferedCaster) PushFileCreate(newFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	fileInfo, err := newFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	msg := wsResponse{
		MessageStatus: "file_created",
		SubscribeKey:  subId(newFile.GetParent().Id()),
		Content:       []map[string]any{gin.H{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushFileUpdate(updatedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	if dataStore.IsSystemDir(updatedFile) {
		return
	}

	fileInfo, err := updatedFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	msg := wsResponse{
		MessageStatus: "file_updated",
		SubscribeKey:  subId(updatedFile.Id()),
		Content:       []map[string]any{gin.H{"fileInfo": fileInfo}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)

	if dataStore.IsSystemDir(updatedFile.GetParent()) {
		return
	}

	msg = wsResponse{
		MessageStatus: "file_updated",
		SubscribeKey:  subId(updatedFile.GetParent().Id()),
		Content:       []map[string]any{gin.H{"fileInfo": fileInfo}},

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

func (c *bufferedCaster) PushFileMove(preMoveFile *dataStore.WeblensFile, postMoveFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	if filepath.Dir(preMoveFile.String()) == filepath.Dir(postMoveFile.String()) {
		util.Error.Println("This should've been a rename")
		return
	}

	c.PushFileCreate(postMoveFile)
	c.PushFileDelete(preMoveFile)
}

func (c *bufferedCaster) PushFileDelete(deletedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	msg := wsResponse{
		MessageStatus: "file_deleted",
		SubscribeKey:  subId(deletedFile.GetParent().Id()),
		Content:       []map[string]any{gin.H{"fileId": deletedFile.Id()}},

		broadcastType: "folder",
	}
	c.bufferAndFlush(msg)
}

func (c *bufferedCaster) PushTaskUpdate(taskId string, status string, result any) {}

func (c *bufferedCaster) Flush() {
	if !c.enabled {
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
	c.bufLock.Unlock()

	if c.autoFlush && len(c.buffer) >= c.bufLimit {
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

	if len(clients) != 0 {
		for _, c := range clients {
			c._writeToClient(r)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", dest.Type, dest.Key)
	}
}
