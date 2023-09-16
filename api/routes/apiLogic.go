package routes

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	"github.com/ethrousseau/weblens/api/util"
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
	Size int `json:"size"`
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
						errContent := map[string]any{"Message": errMsg, "File": util.GuaranteeRelativePath(filepath.Join(path, file["name"].(string)))}
						errorArr = append(errorArr, wsMsg{Type: "error", Content: errContent, Error: "upload_error"})
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
	_, err := b64.URLEncoding.DecodeString(fileHash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.FailOnError(err, "Given filehash (" + fileHash + ") is not base64 encoded")
	}

	includeMeta := util.BoolFromString(ctx.Query("meta"), true)
	includeThumbnail := util.BoolFromString(ctx.Query("thumbnail"), true)
	includeFullres := util.BoolFromString(ctx.Query("fullres"), true)

	if !(includeMeta || includeThumbnail || includeFullres) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.FailOnError(errors.New("at least one of meta, thumbnail, or fullres must be selected"), "Failed to handle get media request (trying to get: " + fileHash + ")")
		// At least one option must be selected
	} else if includeFullres && (includeMeta || includeThumbnail) {
		// Full res must be the only option if selected
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.FailOnError(errors.New("fullres should be the only option if selected"), "Failed to handle get media request (trying to get: " + fileHash + ")")
	}

	db := database.New()
	m := db.GetMedia(fileHash, includeThumbnail)

	if !m.IsFilledOut(!includeThumbnail) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.FailOnError(errors.New("media struct not propperly filled out"), "Failed to get media from Database (trying to get: " + fileHash + ")")
	}

	if includeFullres {
		redisKey := "Fullres " + m.FileHash
		data64, err := db.RedisCacheGet(redisKey)
		if err == nil {
			util.Debug.Println("Redis hit")
			fullResBytes, _ := b64.StdEncoding.DecodeString(data64)
			ctx.Writer.Write(fullResBytes)
			return
		} else {
			util.Debug.Println("Redis miss")
			fullResBytes, err := m.ReadFullres()

			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					ctx.Writer.WriteHeader(http.StatusNotFound)
					return
				} else {
					util.FailOnError(err, "Failed to read full res file")
				}
			}
			ctx.Writer.Write(fullResBytes)
			fullres64 := b64.StdEncoding.EncodeToString(fullResBytes)
			db.RedisCacheSet(redisKey, fullres64)
			return
		}

	} else if !includeMeta && includeThumbnail {

		thumbBytes, err := b64.StdEncoding.DecodeString(m.Thumbnail64)
		util.FailOnError(err, "Failed to decode thumb64 to bytes")

		ctx.Writer.Write(thumbBytes)
	} else {
		ctx.JSON(http.StatusOK, m)
	}

}

func streamVideo(ctx *gin.Context) {
	fileHash := ctx.Param("filehash")
	_, err := b64.URLEncoding.DecodeString(fileHash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		util.FailOnError(err, "Given filehash (" + fileHash + ") is not base64 encoded")
	}

	includeThumbnail := false

	db := database.New()
	m := db.GetMedia(fileHash, includeThumbnail)

	if !m.IsFilledOut(!includeThumbnail) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.FailOnError(errors.New("media struct not propperly filled out"), "Failed to get media from Database (trying to get: " + fileHash + ")")
	}

	// cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-i", util.GuaranteeAbsolutePath(m.Filepath), "-c:v", "h264_videotoolbox", "-f", "h264", "pipe:1")
	// stdout, err := cmd.StdoutPipe()
	// util.FailOnError(err, "Failed to get ffmpeg stdout pipe")

	// util.Debug.Println(cmd.String())

	// err = cmd.Start()
	// if err != nil {
	// 	util.FailOnError(err, "Failed to start ffmpeg")
	// }
	// // buf := new (bytes.Buffer)
	// // _, err = buf.ReadFrom(stdout)

	//	util.FailOnError(err, "Failed to run ffmpeg to get video thumbnail")
	file, err := os.Open(util.GuaranteeAbsolutePath(m.Filepath))

	util.FailOnError(err, "Failed to open fullres stream file")

	ctx.Writer.Header().Add("Connection", "keep-alive")

	//util.Debug.Println(buf)

	//writtenBytes, err := io.Copy(ctx.Writer, stdout)
	writtenBytes, err := io.Copy(ctx.Writer, file)
	util.FailOnError(err, fmt.Sprintf("Failed to write video stream to response writer (wrote %d bytes)", writtenBytes))

	//cmd.Wait()
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
	filepath := filepath.Join(absoluteParent, filename)

	_, err := os.Stat(filepath)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("file (%s) already exists", filename)
	}

	outFile, err := os.Create(filepath)
	util.FailOnError(err, "Failed to create file for uploaded media")

	defer outFile.Close()

	index := strings.Index(item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item64[index + 1:])
	util.FailOnError(err, "Failed to decode uploaded media 64 to bytes")

	_, err = outFile.Write(itemBytes)
	util.FailOnError(err, "Failed to write uploaded bytes to new file")

	db := database.New()

	m, err := importMedia.HandleNewImage(filepath, db)
	if err != nil {
		os.Remove(filepath)
		return nil, err
	}

	return m, nil

}

var dirIgnore = []string{
	".DS_Store",
}

func makeDir(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	absolutePath := filepath.Join(util.GuaranteeAbsolutePath(relativePath), "newDir")

	_, err := os.Stat(absolutePath)
	if !os.IsNotExist(err) {
		util.Error.Panicf("Directory (%s) already exists", relativePath)
	}

	err = os.Mkdir(absolutePath, os.FileMode(0777))
	util.FailOnError(err, "Failed to create new directory")
}

func formatFileInfo(file fs.FileInfo, parentDir string, db database.Weblensdb) (fileInfo, bool) {
	var absolutePath string = util.GuaranteeAbsolutePath(parentDir)
	var relativePath string = util.GuaranteeRelativePath(parentDir)

	var formattedInfo fileInfo
	var include bool = false

	if !slices.Contains(dirIgnore, file.Name()) {
		include = true
		mediaData := db.GetMediaByFilepath(filepath.Join(absolutePath, file.Name()), true)
		filled := mediaData.IsFilledOut(false)
		mediaData.Thumbnail64 = ""

		var fileSize int64
		if file.IsDir() {
			fileSize, _ = util.DirSize(filepath.Join(absolutePath, file.Name()))
		} else {
			fileSize = file.Size()
		}

		formattedInfo = fileInfo{Imported: filled, IsDir: file.IsDir(), Size: int(fileSize), ModTime: file.ModTime(), Filepath: filepath.Join(relativePath, file.Name()), MediaData: mediaData}
	}
	return formattedInfo, include
}

func getDirInfo(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	dirInfo, err := os.ReadDir(absolutePath)
	if err != nil {
		ctx.AbortWithStatus(404)
	}

	db := database.New()
	var filteredDirInfo []fileInfo
	for _, file := range dirInfo {
			info, err := file.Info()
			util.FailOnError(err, "Failed to get file info")
			formattedInfo, include := formatFileInfo(info, absolutePath, db)
			if include {
				filteredDirInfo = append(filteredDirInfo, formattedInfo)
			}
	}

	ctx.JSON(http.StatusOK, filteredDirInfo)
}

func getFile(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	parentDir := filepath.Dir(absolutePath)

	file, _ := os.Stat(absolutePath)
	db := database.New()
	formattedInfo, include := formatFileInfo(file, parentDir, db)

	if include {
		ctx.JSON(http.StatusOK, formattedInfo)
	} else {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}

}

func updateFile(ctx *gin.Context) {
	existingFilepath := util.GuaranteeAbsolutePath(ctx.Query("existingFilepath"))
	newFilepath := util.GuaranteeAbsolutePath(ctx.Query("newFilepath"))

	_, err := os.Stat(existingFilepath)
	util.FailOnError(err, "Cannot rename file that does not exist")

	_, err = os.Stat(newFilepath)
	util.FailOnNoError(err, "Cannot overwrite file that already exists")

	err = os.Rename(existingFilepath, newFilepath)
	util.FailOnError(err, "Failed to rename file")

	db := database.New()
	m := db.GetMediaByFilepath(existingFilepath, true)

	if !m.IsFilledOut(false) {
		util.Error.Panic("Database does not have expected information on existing file")
	}

	m.Filepath = newFilepath
	m.GenerateFileHash()

	db.MoveMedia(existingFilepath, m)

}

func moveFileToTrash(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	absolutePath := util.GuaranteeAbsolutePath(relativePath)

	_, err := os.Stat(absolutePath)
	util.FailOnError(err, "Failed to move file to trash that does not exist")

	trashPath := filepath.Join(util.GetTrashDir(),filepath.Base(absolutePath))

	os.Rename(absolutePath,  trashPath)
	//os.Remove(absolutePath)

	db := database.New()
	db.CreateTrashEntry(relativePath, trashPath)

}