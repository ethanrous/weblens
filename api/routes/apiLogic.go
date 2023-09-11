package routes

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "image/png"

	"slices"

	"github.com/ethrousseau/weblens/api/database"
	"github.com/ethrousseau/weblens/api/importMedia"
	"github.com/ethrousseau/weblens/api/interfaces"
	util "github.com/ethrousseau/weblens/api/utils"
	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

type wsMsg struct {
	Type string 					`json:"type"`
	Content map[string]interface{} 	`json:"content"`
	Error string 					`json:"error"`
}

type fileInfo struct{
	Imported bool `json:"imported"` // If the item has been loaded into the database, dictates if MediaData is set or not
	IsDir bool `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Filepath string `json:"filepath"`
	MediaData interfaces.Media `json:"mediaData"`
}

func wsConnect(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		var msg wsMsg
        err := conn.ReadJSON(&msg)
        if err != nil {
			util.Error.Println(err)
            break
        }
		handleWsRequest(msg, conn)
		conn.WriteJSON(wsMsg{Type: "finished"})
    }
}

func handleWsRequest(msg wsMsg, conn *websocket.Conn) {
	switch msg.Type {
		case "file_upload": {
			var wg sync.WaitGroup

			path := msg.Content["path"].(string)
			files := msg.Content["files"]

			numFiles := len(files.([]any))
			contentArr := make([]fileInfo, numFiles)
			var errorArr []wsMsg

			for i, f := range files.([]any) {
				wg.Add(1)
				go func(file map[string]interface{}, index int) {
					defer wg.Done()
					m, err := uploadItem(path, file["name"].(string), file["item64"].(string))
					if err != nil {
						errMsg := fmt.Sprintf("Upload error: %s", err)
						errorArr = append(errorArr, wsMsg{Type: "error", Error: errMsg})
						return
					}

					f, err := os.Stat(util.GuaranteeAbsolutePath(m.Filepath))
					util.FailOnError(err, "Failed to get stats of uploaded file")

					newItem := fileInfo{
						Imported: true,
						IsDir: false,
						Filepath: util.GuaranteeRelativePath(m.Filepath),
						MediaData: *m,
						ModTime: f.ModTime(),
					}

					contentArr[index] = newItem
				}(f.(map[string]interface{}), i)
			}

			wg.Wait()

			for _, e := range errorArr {
				conn.WriteJSON(e)
			}

			var filteredContentArr []fileInfo

			for _, item := range contentArr {
				if item.Filepath != "" {
					filteredContentArr = append(filteredContentArr, item)
				}
			}

			if len(filteredContentArr) != 0 {
				res := struct {
					Type string 		`json:"type"`
					Content []fileInfo 	`json:"content"`
				} {
					Type: "new_items",
					Content: filteredContentArr,
				}

				conn.WriteJSON(res)
			}
		}
		case "scan_directory": {
			wp := scan(msg.Content["path"].(string), msg.Content["recursive"].(bool))

			_, remainingTasks, totalTasks := wp.Status()
			for remainingTasks > 0 {
				_, remainingTasks, _ = wp.Status()
				status := struct {Type string `json:"type"`; RemainingTasks int `json:"remainingTasks"`; TotalTasks int `json:"totalTasks"`} {Type: "scan_directory_progress", RemainingTasks: remainingTasks, TotalTasks: totalTasks}
				conn.WriteJSON(status)
				time.Sleep(time.Second)
			}
			res := struct {Type string `json:"type"`} {Type: "refresh"}
			conn.WriteJSON(res)
		}
	}
}

func getPagedMedia(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	var err error

	var skip int
	if (ctx.Query("skip") != "") {
		skip, err = strconv.Atoi(ctx.Query("skip"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "skip paramater must be an interger")
			return
		}
	} else {
		skip = 0
	}

	var limit int
	if (ctx.Query("limit") != "") {
		limit, err = strconv.Atoi(ctx.Query("limit"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "limit paramater must be an interger")
			return
		}
	} else {
		limit = 0
	}

	var raw bool
	if ctx.Query("raw") == "" || ctx.Query("raw") == "false" {
		raw = false
	} else {
		raw = true
	}

	db := database.New()
	media, moreMedia := db.GetPagedMedia(sort, skip, limit, raw, true)

	res := struct{
		Media []interfaces.Media
		MoreMedia bool
	} {
		media,
		moreMedia,
	}
	ctx.JSON(http.StatusOK, res)
}

func getMediaItem(ctx *gin.Context) {
	fileHash := ctx.Param("filehash")
	includeMeta := util.BoolFromString(ctx.Query("meta"), true)
	includeThumbnail := util.BoolFromString(ctx.Query("thumbnail"), true)
	includeFullres := util.BoolFromString(ctx.Query("fullres"), true)

	if !(includeMeta || includeThumbnail || includeFullres) {
		// At least one option must be selected
		ctx.AbortWithStatus(http.StatusBadRequest)
	} else if includeFullres && (includeMeta || includeThumbnail) {
		// Full res must be the only option if selected
		ctx.AbortWithStatus(http.StatusBadRequest)
	}

	db := database.New()
	data := db.GetMedia(fileHash, includeThumbnail)

	if includeFullres {
		fullResBytes := data.ReadFullres()
		ctx.Writer.Write(fullResBytes)

	} else if !includeMeta && includeThumbnail {
		thumbBytes, err := b64.StdEncoding.DecodeString(data.Thumbnail64)
		util.FailOnError(err, "Failed to decode thumb64 to bytes")

		ctx.Writer.Write(thumbBytes)
	} else {
		ctx.JSON(http.StatusOK, data)
	}
}

func scan(path string, recursive bool) (util.WorkerPool) {
	scanPath := util.GuaranteeAbsolutePath(path)

	_, err := os.Stat(scanPath)
	util.FailOnError(err, "Scan path does not exist")

	wp := importMedia.ScanDirectory(scanPath, recursive)

	return wp
}

func uploadItem(relParentDir, filename, item64 string) (*interfaces.Media, error) {
	absoluteParent := util.GuaranteeAbsolutePath(relParentDir)
	imgPath := filepath.Join(absoluteParent, filename)

	_, err := os.Stat(imgPath)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("file (%s) already exists", filename)
	}

	outFile, err := os.Create(imgPath)
	util.FailOnError(err, "Failed to create file for uploaded media")

	defer outFile.Close()

	index := strings.Index(item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item64[index + 1:])
	util.FailOnError(err, "Failed to decode uploaded media 64 to bytes")

	_, err = outFile.Write(itemBytes)
	util.FailOnError(err, "Failed to write uploaded bytes to new file")

	db := database.New()

	m, err := importMedia.HandleNewImage(imgPath, db)
	if err != nil {
		return nil, err
	}

	return m, nil

}

var dirIgnore = []string{
	".DS_Store",
}

func makeDir(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	err := os.Mkdir(absolutePath, os.FileMode(0777))
	util.FailOnError(err, "Failed to create new directory")
}

func getDirInfo(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	//go importMedia.ScanDirectory(relativePath, false)

	dirInfo, err := os.ReadDir(absolutePath)
	if err != nil {
		ctx.AbortWithStatus(404)
	}

	db := database.New()
	var filteredDirInfo []fileInfo
	for _, value := range dirInfo {
		info, err := value.Info()
		util.FailOnError(err, "Failed to get file info")
		if !slices.Contains(dirIgnore, value.Name()) {
			mediaData, exists := db.GetMediaByFilepath(filepath.Join(absolutePath, value.Name()))
			filteredDirInfo = append(filteredDirInfo, fileInfo{Imported: exists, IsDir: value.IsDir(), ModTime: info.ModTime(), Filepath: filepath.Join(relativePath, value.Name()), MediaData: mediaData})
		}
	}

	// sort.SliceStable(filteredDirInfo, func(i, j int) bool {
	// 	cmp := filteredDirInfo[i].ModTime.Compare(filteredDirInfo[j].ModTime)
	// 	if cmp == 1 {
	// 		return true
	// 	} else {
	// 		return false
	// 	}
	// })

	ctx.JSON(http.StatusOK, filteredDirInfo)
}

func deleteFile(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	_, err := os.Stat(absolutePath)
	util.FailOnError(err, "Failed to delete file that does not exist")

	os.Remove(absolutePath)

	db := database.New()
	db.RemoveMediaByFilepath(relativePath)

}