package wlcontext

import (
	"context"

	"github.com/ethanrous/weblens/modules/websocket"
)

// Notifier sends WebSocket notifications to connected clients.
type Notifier interface {
	Notify(ctx context.Context, data ...websocket.WsResponseInfo)
}

// NotifierContext provides the ability to send WebSocket notifications.
type NotifierContext interface {
	context.Context
	Notifier
}
