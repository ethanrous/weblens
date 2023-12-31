package routes

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func wsConnect(ctx *gin.Context) {

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	util.FailOnError(err, "Failed to upgrade http request to websocket")

	client := dataProcess.ClientConnect(conn)

	defer client.Disconnect()

	wsUser := ctx.GetString("username")
	util.Info.Printf("%s made successful websocket connection (%s)", wsUser, client.GetClientId())

	for {
		_, buf, err := conn.ReadMessage()
        if err != nil {
            break
        }
		go handleWsRequest(buf, client, wsUser)
    }
}

func handleWsRequest(msgBuf []byte, client *dataProcess.Client, wsUser string) {
	var msg dataProcess.WsRequest
	err := json.Unmarshal(msgBuf, &msg)
	util.FailOnError(err, "Failed to unmarshal ws message")

	wsReqSwitchboard(msg, client, wsUser)
}

func wsReqSwitchboard(msg dataProcess.WsRequest, client *dataProcess.Client, username string) {
	defer util.RecoverPanic("[WEBSOCKET] Client %d panicked", client.GetClientId())

	fmt.Printf("[WEBSOCKET] %s | %s | %s\n", time.Now().Format("2006/01/02 - 15:04:05"), client.GetClientId(), msg.ReqType)

	switch msg.ReqType {
		case "subscribe": {
			subType, meta := getSubscribeInfo(msg.Content.(map[string]any))
			if subType == "" || meta == nil {
				util.Error.Printf("Bad subscribe request: %v", msg.Content)
				return
			}
			complete, result := client.Subscribe(subType, username, meta)
			if complete {
				client.Send("zip_complete", struct {TakeoutId string `json:"takeoutId"`} {TakeoutId: result}, nil)
			}
		}

		case "scan_directory": {
			var scanInfo dataProcess.ScanContent
			util.StructFromMap(msg.Content.(map[string]any), &scanInfo)

			folder := dataStore.WFDByFolderId(scanInfo.FolderId)
			util.FailOnError(folder.Err(), "Failed to get folder to scan")

			meta := dataProcess.ScanMetadata{File: folder, Username: username, Recursive: scanInfo.Recursive}
			dataProcess.RequestTask("scan_directory", "", meta)
		}

		default: {
			util.Error.Printf("Could not parse websocket request type: %v", msg)
		}
	}
}

func getSubscribeInfo(contentMap map[string]any) (string, any) {
	subType := contentMap["subType"].(string)
	delete(contentMap, "subType")
	switch subType {
	case "folder": {
		var subContentStruct dataProcess.FolderSubMetadata
		err := util.StructFromMap(contentMap, &subContentStruct)
		util.FailOnError(err, "Could not convert map to struct")
		return subType, subContentStruct
	}
	case "task": {
		var taskContentStruct dataProcess.TaskSubMetadata
		err := util.StructFromMap(contentMap, &taskContentStruct)
		util.FailOnError(err, "Could not convert map to struct")
		return subType, taskContentStruct
	}
	default: {
		util.Error.Printf("Unknown subscribe type: [ %s ] -- RAW: [ %v ]", subType, contentMap)
	}
	}
	return "", nil
}