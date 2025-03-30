package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/websocket"
)

type Notifier interface {
	Notify(data ...websocket.WsResponseInfo)
}

type NotifierContext interface {
	context.Context
	Notifier
}
