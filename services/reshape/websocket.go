package reshape

import (
	"time"

	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
)

func getSafeString(content websocket_mod.WsData, key string) (val string) {
	if content != nil {
		if id, ok := content[key]; ok {
			val = id.(string)
		}
	}

	return
}

// GetSubscribeInfo extracts subscription information from a websocket response message.
func GetSubscribeInfo(msg websocket_mod.WsResponseInfo) websocket_mod.SubscriptionInfo {
	return websocket_mod.SubscriptionInfo{
		When:           time.UnixMilli(msg.SentTime),
		Type:           msg.BroadcastType,
		SubscriptionID: msg.SubscribeKey,
		ShareID:        getSafeString(msg.Content, "shareID"),
	}
}

// GetCancelInfo extracts cancellation information from a websocket response message.
func GetCancelInfo(msg websocket_mod.WsResponseInfo) websocket_mod.CancelInfo {
	return websocket_mod.CancelInfo{
		TaskID: msg.Content["taskID"].(string),
	}
}

// GetScanInfo extracts scan information from a websocket response message.
func GetScanInfo(msg websocket_mod.WsResponseInfo) websocket_mod.ScanInfo {
	id := getSafeString(msg.Content, "folderID")
	if id == "" {
		id = getSafeString(msg.Content, "fileID")
	}

	return websocket_mod.ScanInfo{
		FileID:  id,
		ShareID: getSafeString(msg.Content, "shareID"),
	}
}
