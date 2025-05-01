package reshape

import (
	"time"

	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
)

func getSafeString(content websocket_mod.WsData, key string) (val string) {
	if content != nil {
		if id, ok := content["shareId"]; ok {
			val = id.(string)
		}
	}

	return
}

func GetSubscribeInfo(msg websocket_mod.WsResponseInfo) websocket_mod.SubscriptionInfo {
	return websocket_mod.SubscriptionInfo{
		When:           time.UnixMilli(msg.SentTime),
		Type:           msg.BroadcastType,
		SubscriptionId: msg.SubscribeKey,
		ShareId:        getSafeString(msg.Content, "shareId"),
	}
}

func GetCancelInfo(msg websocket_mod.WsResponseInfo) websocket_mod.CancelInfo {
	return websocket_mod.CancelInfo{
		TaskId: msg.SubscribeKey,
	}
}

func GetScanInfo(msg websocket_mod.WsResponseInfo) websocket_mod.ScanInfo {
	return websocket_mod.ScanInfo{
		FileId:  getSafeString(msg.Content, "fileId"),
		ShareId: getSafeString(msg.Content, "shareId"),
	}
}
