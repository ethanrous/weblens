package routes

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func wsConnect(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	util.Debug.Println("Websocket connected")

	for {
		var msg wsMsg
        err := conn.ReadJSON(&msg)
        if err != nil {
			util.Error.Println(err)
            break
        }
		go wsReqSwitchboard(msg, conn)
		conn.WriteJSON(wsMsg{Type: "finished"})
    }
}

func wsReqSwitchboard(msg wsMsg, conn *websocket.Conn) {
	switch msg.Type {
		case "file_upload": {

			relPath := dataStore.GuaranteeRelativePath(msg.Content["path"].(string))

			fileData := msg.Content["file"].(map[string]interface {})
			m, err := uploadItem(relPath, fileData["name"].(string), fileData["item64"].(string))

			if err != nil {
				errMsg := fmt.Sprintf("Upload error: %s", err)
				errContent := map[string]any{"Message": errMsg, "File": dataStore.GuaranteeRelativePath(filepath.Join(relPath, fileData["name"].(string)))}
				conn.WriteJSON(wsMsg{Type: "error", Content: errContent, Error: "upload_error"})
				return
			}

			f, err := os.Stat(dataStore.GuaranteeAbsolutePath(m.Filepath))
			util.FailOnError(err, "Failed to get stats of uploaded file")

			newItem := fileInfo{
				Imported: true,
				IsDir: false,
				Size: int(f.Size()),
				Filepath: dataStore.GuaranteeRelativePath(m.Filepath),
				MediaData: *m,
				ModTime: f.ModTime(),
			}

			res := struct{
				Type string 		`json:"type"`
				Content []fileInfo 	`json:"content"`
			} {
				Type: "new_items",
				Content: []fileInfo{newItem},
			}

			conn.WriteJSON(res)
		}

		case "scan_directory": {
			wp := scan(msg.Content["path"].(string), msg.Content["recursive"].(bool))

			var previousRemaining int
			_, remainingTasks, totalTasks := wp.Status()
			for remainingTasks > 0 {
				time.Sleep(time.Second)
				_, remainingTasks, _ = wp.Status()

				// Don't send new message unless new data
				if remainingTasks == previousRemaining {
					continue
				} else {
					previousRemaining = remainingTasks
				}

				status := struct {Type string `json:"type"`; RemainingTasks int `json:"remainingTasks"`; TotalTasks int `json:"totalTasks"`} {Type: "scan_directory_progress", RemainingTasks: remainingTasks, TotalTasks: totalTasks}
				conn.WriteJSON(status)
			}
			res := struct {Type string `json:"type"`} {Type: "refresh"}
			conn.WriteJSON(res)
		}
	}
}