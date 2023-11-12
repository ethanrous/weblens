package routes

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
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
		var msg dataProcess.WsMsg
		_, buf, err := conn.ReadMessage()
        // err := conn.ReadJSON(&msg)
        if err != nil {
			util.Error.Println(err)
            break
        }
		err = json.Unmarshal(buf, &msg)
		util.FailOnError(err, "Failed to unmarshal ws message")

		go wsReqSwitchboard(msg, client, wsUser)
		conn.WriteJSON(dataProcess.WsMsg{Type: "finished"})
    }
}

func wsReqSwitchboard(msg dataProcess.WsMsg, client *dataProcess.Client, username string) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("[WEBSOCKET] Recovered panic while handling message (%s): %v\n", msg.Type, err)
			// util.DisplayError(err, "[WEBSOCKET] Recovered panic while handling message")
			fmt.Printf("[WEBSOCKET] %s | %s | %s\n", time.Now().Format("2006/01/02 - 15:04:05"), client.GetClientId(), msg.Type)
		}
	}()

	fmt.Printf("[WEBSOCKET] %s | %s | %s\n", time.Now().Format("2006/01/02 - 15:04:05"), client.GetClientId(), msg.Type)

	jsonString, err := json.Marshal(msg.Content)
	util.FailOnError(err, "Failed to marshal ws content to json string")
	switch msg.Type {
		case "subscribe": {
			var subMeta dataProcess.SubscribeContent
			json.Unmarshal(jsonString, &subMeta)
			client.Subscribe(subMeta, username)
		}

		case "scan_directory": {
			var content dataProcess.ScanContent
			json.Unmarshal(jsonString, &content)

			path := content.Path
			recursive := content.Recursive

			meta := dataProcess.ScanMetadata{Path: path, Username: username, Recursive: recursive}
			metaS, _ := json.Marshal(meta)

			task := dataProcess.Task{TaskType: "scan_directory", Metadata: string(metaS)}
			dataProcess.RequestTask(task)

		}

		default: {
			util.Error.Printf("Recieved unknown websocket message type: %s", msg.Type)
		}
	}
}