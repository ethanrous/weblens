package routes

import "github.com/ethrousseau/weblens/api/dataStore"

// Ws response types
type WsResponse struct {
	MessageStatus string `json:"messageStatus"`
	SubscribeKey  string `json:"subscribeKey"`
	Content       any    `json:"content"`
	Error         string `json:"error"`
}

// Physical type to pass BroadcasterAgent to children
type caster struct {
	enabled bool
}

var Caster *caster = &caster{enabled: false}

func (c *caster) Enable() {
	c.enabled = true
}

type BroadcasterAgent interface {
	PushItemCreate(newFile *dataStore.WeblensFileDescriptor)
	PushItemUpdate(updatedFile *dataStore.WeblensFileDescriptor)
	PushItemMove(preMoveFile *dataStore.WeblensFileDescriptor, postMoveFile *dataStore.WeblensFileDescriptor)
	PushItemDelete(deletedFile *dataStore.WeblensFileDescriptor)

	PushTaskUpdate(taskId string, status string, result any)
}
