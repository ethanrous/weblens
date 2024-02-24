package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func wsConnect(ctx *gin.Context) {

	ctx.Status(http.StatusSwitchingProtocols)
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		util.DisplayError(err)
		return
	}

	client := cmInstance.ClientConnect(conn, ctx.GetString("username"))
	go wsMain(client)
}

func wsMain(client *Client) {
	defer client.Disconnect()

	for {
		_, buf, err := client.conn.ReadMessage()
		if err != nil {
			break
		}
		go wsReqSwitchboard(buf, client)
	}
}

func wsReqSwitchboard(msgBuf []byte, client *Client) {
	defer wsRecover(client.GetClientId())
	// defer util.RecoverPanic("[WS] Client %d panicked: %v", client.GetClientId())

	var msg wsRequest
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		util.DisplayError(err)
		return
	}

	switch msg.Action {
	case Subscribe:
		{
			var subInfo subscribeInfo
			err = json.Unmarshal([]byte(msg.Content), &subInfo)

			if err != nil {
				util.DisplayError(err)
				client.Error(errors.New("failed to parse subscribe request"))
			}

			if subInfo.SubType == "" || subInfo.Key == "" {
				client.Error(fmt.Errorf("bad subscribe request: %s", msg.Content))
				return
			}

			complete, result := client.Subscribe(subInfo.SubType, subInfo.Key, subInfo.Meta)
			if complete {
				Caster.PushTaskUpdate(string(subInfo.Key), "zip_complete", gin.H{"takeoutId": result})
			}
		}

	case Unsubscribe:
		{
			var unsubInfo unsubscribeInfo
			json.Unmarshal([]byte(msg.Content), &unsubInfo)
			client.Unsubscribe(unsubInfo.Key)
		}

	case ScanDirectory:
		{
			var scanInfo scanInfo
			err := json.Unmarshal([]byte(msg.Content), &scanInfo)
			if err != nil {
				util.DisplayError(err)
				return
			}
			folder := dataStore.FsTreeGet(scanInfo.FolderId)
			if folder == nil {
				util.Error.Println("Failed to get dir to scan:", scanInfo.FolderId)
				return
			}

			t := dataProcess.GetGlobalQueue().ScanDirectory(folder, scanInfo.Recursive, scanInfo.DeepScan, Caster)
			client.Subscribe(SubTask, subId(t.TaskId()), nil)
		}

	default:
		{
			client.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}
}

func wsRecover(clientId string) {
	err := recover()
	if err != nil {
		util.WsError.Println(clientId, err, debug.Stack())
	}
}
