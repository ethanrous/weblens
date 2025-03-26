package context

import "github.com/ethanrous/weblens/modules/websocket"

type Notifier interface {
	Notify(message string)
}

type NotifierContext struct {
	BasicContext

	notifier Notifier
}

func NewNotifierContext(ctx BasicContext, notifier Notifier) NotifierContext {
	return NotifierContext{
		BasicContext: ctx,
		notifier:     notifier,
	}
}

func (c *NotifierContext) Notify(event websocket.WsEvent, message string) {
	c.notifier.Notify(message)
}
