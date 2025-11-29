package websocket

import (
	"time"
)

type WsData map[string]any

type WsResponseInfo struct {
	Action          WsAction         `json:"action"`
	EventTag        WsEvent          `json:"eventTag"`
	SubscribeKey    string           `json:"subscribeKey"`
	TaskType        string           `json:"taskType,omitempty"`
	Content         WsData           `json:"content"`
	Error           string           `json:"error,omitempty"`
	BroadcastType   SubscriptionType `json:"subscriptionType,omitempty"`
	RelaySource     string           `json:"relaySource,omitempty"`
	SentTime        int64            `json:"sentTime,omitempty"`
	ConstructedTime int64            `json:"constructedTime,omitempty"`

	Sent chan struct{} `json:"-"`
}

// type WsRequestInfo struct {
// 	Content string `json:"content"`
// 	SentAt  int64  `json:"sentAt"`
// }

// WsR Request interface
type WsR interface {
	GetKey() string
	Action() WsAction
	GetShareId() string
}

type Subscription struct {
	When           time.Time
	Type           SubscriptionType
	SubscriptionId string
}

type SubscriptionInfo struct {
	When           time.Time
	Type           SubscriptionType
	SubscriptionId string
	ShareId        string
}

type ScanInfo struct {
	FileId  string
	ShareId string
}

type CancelInfo struct {
	TaskId string
}
