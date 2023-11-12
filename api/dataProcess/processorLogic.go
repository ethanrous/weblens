package dataProcess

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func _scan(path, username string, recursive bool) (WorkerPool) {
	scanPath := dataStore.GuaranteeUserAbsolutePath(path, username)

	_, err := os.Stat(scanPath)
	util.FailOnError(err, "Scan path does not exist")

	wp := ScanDirectory(scanPath, username, recursive)

	return wp
}

type ScanMetadata struct {
	Path string
	Username string
	Recursive bool
}

// path, username string, recursive bool
func Scan(metaS string) {
	var meta ScanMetadata
	json.Unmarshal([]byte(metaS), &meta)

	absolutePath := dataStore.GuaranteeUserAbsolutePath(meta.Path, meta.Username)
	wp := _scan(meta.Path, meta.Username, meta.Recursive)

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
		Broadcast(absolutePath, status)

	}
	// res := struct {Type string `json:"type"`} {Type: "refresh"}
	// Broadcast(absolutePath, res)
}

func PushItemUpdate(p, username string, db dataStore.Weblensdb) {
	fileInfo, _ := dataStore.FormatFileInfo(p, username, db)

	msg := WsMsg{Type: "item_update", Content: fileInfo}

	Broadcast(filepath.Dir(p), msg)
}