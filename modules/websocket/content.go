// Package websocket provides WebSocket communication functionality for real-time client-server messaging.
package websocket

import (
	"time"
)

// WsData represents a map of key-value pairs for WebSocket message content.
type WsData map[string]any

// WsResponseInfo represents a WebSocket response message sent to clients.
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
	GetShareID() string
}

// Subscription represents a client's subscription to WebSocket events.
type Subscription struct {
	When           time.Time
	Type           SubscriptionType
	SubscriptionID string
}

// SubscriptionInfo represents detailed information about a client's subscription including share context.
type SubscriptionInfo struct {
	When           time.Time
	Type           SubscriptionType
	SubscriptionID string
	ShareID        string
}

// ScanInfo represents information about a file scanning operation.
type ScanInfo struct {
	FileID  string
	ShareID string
}

// CancelInfo represents information for canceling a running task.
type CancelInfo struct {
	TaskID string
}
