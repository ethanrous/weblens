package routes

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "image/png"

	"slices"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/importMedia"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"

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
	MediaData dataStore.Media `json:"mediaData"`
}

type loginInfo struct {
	Username string 	`json:"username"`
	Password string 	`json:"password"`
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

	db := dataStore.NewDB()
	media, moreMedia := db.GetPagedMedia(sort, skip, limit, raw, true)

	res := struct{
		Media []dataStore.Media
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
		util.DisplayError(err, "Given filehash (" + fileHash + ") is not base64 encoded")
		return
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

	db := dataStore.NewDB()
	m := db.GetMedia(fileHash, includeThumbnail)

	if !m.IsFilledOut(!includeThumbnail) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.FailOnError(errors.New("media struct not propperly filled out"), "Failed to get media from Database (trying to get: " + fileHash + ")")
	}

	if includeFullres {
		fullresBytes, err := m.ReadFullres(db)
		util.FailOnError(err, "Failed to read fullres file")
		ctx.Writer.Write(fullresBytes)

		return
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
		util.DisplayError(err, "Given filehash (" + fileHash + ") is not base64 encoded")
		return
	}

	includeThumbnail := false

	db := dataStore.NewDB()
	m := db.GetMedia(fileHash, includeThumbnail)

	if !m.IsFilledOut(!includeThumbnail) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.FailOnError(errors.New("media struct not propperly filled out"), "Failed to get media from Database (trying to get: " + fileHash + ")")
	}

	// cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-i", dataStore.GuaranteeAbsolutePath(m.Filepath), "-c:v", "h264_videotoolbox", "-f", "h264", "pipe:1")
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
	file, err := os.Open(dataStore.GuaranteeAbsolutePath(m.Filepath))

	util.FailOnError(err, "Failed to open fullres stream file")

	ctx.Writer.Header().Add("Connection", "keep-alive")

	//util.Debug.Println(buf)

	//writtenBytes, err := io.Copy(ctx.Writer, stdout)
	writtenBytes, err := io.Copy(ctx.Writer, file)
	util.FailOnError(err, fmt.Sprintf("Failed to write video stream to response writer (wrote %d bytes)", writtenBytes))

	//cmd.Wait()
}

func scan(path string, recursive bool) (util.WorkerPool) {
	scanPath := dataStore.GuaranteeAbsolutePath(path)

	_, err := os.Stat(scanPath)
	util.FailOnError(err, "Scan path does not exist")

	wp := importMedia.ScanDirectory(scanPath, recursive)

	return wp
}

func uploadItem(relParentDir, filename, item64 string) (*dataStore.Media, error) {
	absoluteParent := dataStore.GuaranteeAbsolutePath(relParentDir)
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

	db := dataStore.NewDB()

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
	absolutePath := filepath.Join(dataStore.GuaranteeAbsolutePath(relativePath), "untitled folder")

	_, err := os.Stat(absolutePath)

	counter := 1
	for err == nil {
		absolutePath = filepath.Join(dataStore.GuaranteeAbsolutePath(relativePath), fmt.Sprintf("untitled folder %d", counter))
		_, err = os.Stat(absolutePath)
		counter += 1
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		util.Error.Panicf("Directory (%s) already exists", relativePath)
	}

	err = os.Mkdir(absolutePath, os.FileMode(0777))
	util.FailOnError(err, "Failed to create new directory")

	ctx.JSON(http.StatusCreated, dataStore.GuaranteeRelativePath(absolutePath))
}

func formatFileInfo(file fs.FileInfo, parentDir string, db dataStore.Weblensdb) (fileInfo, bool) {
	var absolutePath string = dataStore.GuaranteeAbsolutePath(parentDir)
	var relativePath string = dataStore.GuaranteeRelativePath(parentDir)

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
	absolutePath := dataStore.GuaranteeAbsolutePath(relativePath)

	dirInfo, err := os.ReadDir(absolutePath)
	if err != nil {
		ctx.AbortWithStatus(404)
		return
	}

	db := dataStore.NewDB()
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
	absolutePath := dataStore.GuaranteeAbsolutePath(relativePath)

	parentDir := filepath.Dir(absolutePath)

	file, _ := os.Stat(absolutePath)
	db := dataStore.NewDB()
	formattedInfo, include := formatFileInfo(file, parentDir, db)

	if include {
		ctx.JSON(http.StatusOK, formattedInfo)
	} else {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}

}

func updateFile(ctx *gin.Context) {
	existingFilepath := dataStore.GuaranteeAbsolutePath(ctx.Query("existingFilepath"))
	newFilepath := dataStore.GuaranteeAbsolutePath(ctx.Query("newFilepath"))

	stat, err := os.Stat(existingFilepath)
	util.FailOnError(err, "Cannot rename file that does not exist")

	_, err = os.Stat(newFilepath)
	util.FailOnNoError(err, "Cannot overwrite file that already exists")

	err = os.Rename(existingFilepath, newFilepath)
	util.FailOnError(err, "Failed to rename file")

	if stat.IsDir() {
		return
	}

	db := dataStore.NewDB()
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
	absolutePath := dataStore.GuaranteeAbsolutePath(relativePath)

	_, err := os.Stat(absolutePath)
	util.FailOnError(err, "Failed to move file to trash that does not exist")

	trashPath := filepath.Join(util.GetTrashDir(),filepath.Base(absolutePath))

	_, err = os.Stat(trashPath)
	if err == nil {
		trashPath += time.Now().Format(" 2006-01-02 15.04.05")
	}
	err = os.Rename(absolutePath,  trashPath)
	util.FailOnError(err, "Failed to move file to trash")

	db := dataStore.NewDB()
	db.CreateTrashEntry(relativePath, trashPath)
}

type takeoutItems struct {
	Items []string `json:"items"`
	Path []string `json:"path"`
}

func createTakeout(ctx *gin.Context) {
	util.Debug.Println("Beginning takeout creation")

	var items takeoutItems
	bodyData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Could not read items to create takeout with")
	json.Unmarshal(bodyData, &items)

	var redir string
	var doSingle bool = false

	if len(items.Items) == 1 {
		stat, _ := os.Stat(dataStore.GuaranteeAbsolutePath(items.Items[0]))
		if !stat.IsDir() {
			doSingle = true
		}
	}

	if doSingle {
		db := dataStore.NewDB()
		m := db.GetMediaByFilepath(items.Items[0], false)
		redir = fmt.Sprintf("/api/takeout/%s?single=true", m.FileHash)

	} else {
		takeoutId := dataStore.CreateZipFromPaths(items.Items)
		redir = fmt.Sprintf("/api/takeout/%s", takeoutId)
	}

	ctx.Redirect(http.StatusSeeOther, redir)

}

func getTakeout(ctx *gin.Context) {
	takeoutId := ctx.Param("takeoutId")
	isSingle := util.BoolFromString(ctx.Query("single"), true)

	extraHeaders := make(map[string]string)

	var readPath string

	if isSingle {
		db := dataStore.NewDB()
		m := db.GetMedia(takeoutId, false)
		readPath = dataStore.GuaranteeAbsolutePath(m.Filepath)
		extraHeaders["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\";", filepath.Base(m.Filepath))
	} else {
		readPath = fmt.Sprintf("%s/%s.zip", util.GetTakeoutDir(), takeoutId) // Will already be absolute path
		extraHeaders["Content-Disposition"] = "attachment; filename=\"takeout.zip\";"
	}

	f, err := os.Open(readPath)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		util.FailOnError(err, "Failed to open takeout file")
	}

	stat, err := f.Stat()
	util.FailOnError(err, "")

	extraHeaders["Access-Control-Expose-Headers"] = "Content-Disposition"
	ctx.DataFromReader(http.StatusOK, stat.Size(), "application/octet-stream", f, extraHeaders)

}

type tokenReturn struct {
	Token string `json:"token"`
}

func loginUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var usrCreds loginInfo
	json.Unmarshal(jsonData, &usrCreds)

	//passHashBytes, err := bcrypt.GenerateFromPassword([]byte(usrCreds.Password), 14)
	if err != nil {
		panic(err)
	}
	//passHash := string(passHashBytes)

	db := dataStore.NewDB()
	if db.CheckLogin(usrCreds.Username, usrCreds.Password) {
		util.Debug.Printf("Valid login for [%s]\n", usrCreds.Username)

		token := jwt.New(jwt.SigningMethodHS256)
		tokenString, err := token.SignedString([]byte("key"))
		if err != nil {
			util.Error.Println(err)
		}

		util.Debug.Println(tokenString)
		var ret tokenReturn = tokenReturn{Token: tokenString}

		db.AddTokenToUser(usrCreds.Username, tokenString)

		ctx.JSON(http.StatusOK, ret)
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}


}