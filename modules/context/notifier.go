package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/websocket"
)

type Notifier interface {
	Notify(ctx context.Context, data ...websocket.WsResponseInfo)
}

type NotifierContext interface {
	context.Context
	Notifier
}
