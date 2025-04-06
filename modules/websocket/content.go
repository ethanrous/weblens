package websocket

import (
	"time"
)

type WsData map[string]any

type WsResponseInfo struct {
	EventTag        WsEvent  `json:"eventTag"`
	SubscribeKey    string   `json:"subscribeKey"`
	TaskType        string   `json:"taskType,omitempty"`
	Content         WsData   `json:"content"`
	Error           string   `json:"error,omitempty"`
	BroadcastType   WsAction `json:"broadcastType,omitempty"`
	RelaySource     string   `json:"relaySource,omitempty"`
	SentTime        int64    `json:"sentTime,omitempty"`
	ConstructedTime int64    `json:"constructedTime,omitempty"`
}

type WsRequestInfo struct {
	Action  WsAction `json:"action"`
	Content string   `json:"content"`
	SentAt  int64    `json:"sentAt"`
}

// WsR Request interface
type WsR interface {
	GetKey() string
	Action() WsAction
	GetShareId() string
}

type Subscription struct {
	When           time.Time
	Type           WsAction
	SubscriptionId string
}
