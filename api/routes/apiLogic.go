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

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

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

	db := dataStore.NewDB(ctx.GetString("username"))
	user, err := db.GetUser(ctx.GetString("username"))
	util.FailOnError(err, "Failed to get user for paged media")

	media, moreMedia := db.GetPagedMedia(sort, user.Username, skip, limit, raw, true)

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

	db := dataStore.NewDB(ctx.GetString("username"))
	m := db.GetMedia(fileHash, includeThumbnail)

	filled, reason := m.IsFilledOut(!includeThumbnail)
	if !filled {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.Error.Printf("Failed to get [ %s ] from Database (missing %s)", fileHash, reason)
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

type updateMediaItemsBody struct {
	Owner string 		`json:"owner"`
	FileHashes []string `json:"fileHashes"`
}

func updateMediaItems(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read body of media item update")

	var updateBody updateMediaItemsBody
	err = json.Unmarshal(jsonData, &updateBody)
	util.FailOnError(err, "Failed to unmarshal body of media item update")
	// util.Debug.Println(bytes.NewBuffer(jsonData).String())

	db := dataStore.NewDB(ctx.GetString("username"))
	db.UpdateMediasByFilehash(updateBody.FileHashes, updateBody.Owner)

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

	db := dataStore.NewDB(ctx.GetString("username"))
	m := db.GetMedia(fileHash, includeThumbnail)

	filled, reason := m.IsFilledOut(!includeThumbnail)
	if !filled {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.Error.Printf("Failed to get [ %s ] from Database (missing %s)", fileHash, reason)
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
	file, err := os.Open(dataStore.GuaranteeAbsolutePath(m.Filepath, ctx.GetString("username")))

	util.FailOnError(err, "Failed to open fullres stream file")

	ctx.Writer.Header().Add("Connection", "keep-alive")

	//util.Debug.Println(buf)

	//writtenBytes, err := io.Copy(ctx.Writer, stdout)
	writtenBytes, err := io.Copy(ctx.Writer, file)
	util.FailOnError(err, fmt.Sprintf("Failed to write video stream to response writer (wrote %d bytes)", writtenBytes))

	//cmd.Wait()
}



func uploadItem(relParentDir, filename, item64, uploaderName string) {
	absoluteParent := dataStore.GuaranteeUserAbsolutePath(relParentDir, uploaderName)
	itemPath := filepath.Join(absoluteParent, filename)

	_, err := os.Stat(itemPath)
	if !os.IsNotExist(err) {
		newErr := fmt.Errorf("file (%s) already exists", filename)
		util.DisplayError(newErr, "")
	}

	outFile, err := os.Create(itemPath)
	util.FailOnError(err, "Failed to create file for uploaded media")

	defer outFile.Close()

	index := strings.Index(item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item64[index + 1:])
	util.FailOnError(err, "Failed to decode uploaded media 64 to bytes")

	_, err = outFile.Write(itemBytes)
	util.FailOnError(err, "Failed to write uploaded bytes to new file")

	// db := dataStore.NewDB(uploaderName)

	// uploader, err := db.GetUser(uploaderName)
	// util.FailOnError(err, "Failed to get uploader user")


	util.Debug.Println("ASKING TO SCAN")
	dataProcess.RequestTask("scan_file", dataProcess.ScanMetadata{Path: itemPath, Username: uploaderName})
	// m, err := dataProcess.HandleNewImage(itemPath, uploaderName, db)
	// if err != nil {
		// os.Remove(itemPath)
		// return nil, err
	// }

	// dataProcess.PushItemUpdate(itemPath, uploaderName, db)
	// newItem, _ := dataStore.FormatFileInfo(itemPath, uploaderName, db)

	// msg := dataProcess.WsMsg{Type: "item_update", Content: newItem}
	// dataProcess.Broadcast(filepath.Dir(itemPath), msg)

	// return m, nil
}

type uploadedFile struct {
	File64 string `json:"file64"`
	FileName string `json:"fileName"`
	Path string `json:"path"`
}

func uploadFile(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	// util.Debug.Println(jsonData)
	util.FailOnError(err, "Failed to read request body on file upload")

	var file uploadedFile
	err = json.Unmarshal(jsonData, &file)
	util.FailOnError(err, "Failed to unmarshal request body in file upload")

	username := ctx.GetString("username")
	uploadItem(file.Path, file.FileName, file.File64, username)
}

func makeDir(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	username := ctx.GetString("username")
	absolutePath := filepath.Join(dataStore.GuaranteeUserAbsolutePath(relativePath, username), "untitled folder")

	_, err := os.Stat(absolutePath)

	counter := 1
	for err == nil {
		absolutePath = filepath.Join(dataStore.GuaranteeAbsolutePath(relativePath, username), fmt.Sprintf("untitled folder %d", counter))
		_, err = os.Stat(absolutePath)
		counter += 1
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		util.Error.Panicf("Directory (%s) already exists", relativePath)
	}

	err = os.Mkdir(absolutePath, os.FileMode(0777))
	util.FailOnError(err, "Failed to create new directory")

	userPath, _ := dataStore.GuaranteeUserRelativePath(absolutePath, username)
	ctx.JSON(http.StatusCreated, userPath)

	db := dataStore.NewDB(username)

	dataProcess.PushItemUpdate(absolutePath, username, db)
	// info, _ := dataStore.FormatFileInfo(absolutePath, username, db)
	// msg := dataProcess.WsMsg{Type: "item_update", Content: info}
	// dataProcess.Broadcast(absolutePath, msg)
}

func createUserHomeDir(username string) {
	homeDirPath := dataStore.GuaranteeAbsolutePath("/", username)
	_, err := os.Stat(homeDirPath)
	if err == nil {
		util.Error.Panicln("Tried to create user home directory, but it already exists")
	}
	os.Mkdir(homeDirPath, os.FileMode(0777))
}

func getDirInfo(ctx *gin.Context) () {
	relativePath := ctx.Query("path")
	username := ctx.GetString("username")
	absolutePath := dataStore.GuaranteeUserAbsolutePath(relativePath, username)
	db := dataStore.NewDB(username)

	cachedJson, err := db.RedisCacheGet(absolutePath)
	if err == nil {
		ctx.Data(http.StatusOK, "application/json", []byte(cachedJson))
		return
	}

	dirInfo, err := os.ReadDir(absolutePath)
	if err != nil {
		ctx.AbortWithStatus(404)
		return
	}

	var filteredDirInfo []dataStore.FileInfo
	for _, file := range dirInfo {
		formattedInfo, include := dataStore.FormatFileInfo(filepath.Join(absolutePath, file.Name()), username, db)
		if include {
			filteredDirInfo = append(filteredDirInfo, formattedInfo)
		}
	}

	ctx.JSON(http.StatusOK, filteredDirInfo)

	dirInfoJson, err := json.Marshal(filteredDirInfo)
	util.FailOnError(err, "Failed to marshal dir info to json string")
	db.RedisCacheSet(absolutePath, string(dirInfoJson))
}

func getFile(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	username := ctx.GetString("username")
	absolutePath := dataStore.GuaranteeAbsolutePath(relativePath, username)

	db := dataStore.NewDB(username)
	formattedInfo, include := dataStore.FormatFileInfo(absolutePath, username, db)

	if include {
		ctx.JSON(http.StatusOK, formattedInfo)
	} else {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}

}

func updateFile(ctx *gin.Context) {
	username := ctx.GetString("username")
	existingFilepath := dataStore.GuaranteeUserAbsolutePath(ctx.Query("existingFilepath"), username)
	newFilepath := dataStore.GuaranteeUserAbsolutePath(ctx.Query("newFilepath"), username)

	stat, err := os.Stat(existingFilepath)
	util.FailOnError(err, "Cannot rename file that does not exist")

	_, err = os.Stat(newFilepath)
	util.FailOnNoError(err, "Cannot overwrite file that already exists")

	err = os.Rename(existingFilepath, newFilepath)
	util.FailOnError(err, "Failed to rename file")

	if stat.IsDir() {
		return
	}

	db := dataStore.NewDB(username)
	m, _ := db.GetMediaByFilepath(existingFilepath, true)

	filled, reason := m.IsFilledOut(!false)
	if !filled {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		util.Error.Printf("Failed to get [ %s ] from Database (missing %s)", existingFilepath, reason)
	}

	db.MoveMedia(newFilepath, username, m)
	db.RedisCacheBust(filepath.Dir(newFilepath))

	dataProcess.PushItemUpdate(filepath.Dir(newFilepath), username, db)

}

func moveFileToTrash(ctx *gin.Context) {
	relativePath := ctx.Query("path")
	username := ctx.GetString("username")
	absolutePath := dataStore.GuaranteeUserAbsolutePath(relativePath, username)

	file, err := os.Stat(absolutePath)
	util.FailOnError(err, "Failed to move file to trash that does not exist")

	trashPath := filepath.Join(util.GetTrashDir(),filepath.Base(absolutePath))

	_, err = os.Stat(trashPath)
	if err == nil {
		trashPath += time.Now().Format(" 2006-01-02 15.04.05")
	}

	db := dataStore.NewDB(username)
	var m dataStore.Media
	if file.IsDir() {
		filepath.WalkDir(absolutePath, func (path string, d fs.DirEntry, err error) error {
			if path == absolutePath {
				return nil
			}
		 	relPath, _ := dataStore.GuaranteeUserRelativePath(path, username)
			m, _ := db.GetMediaByFilepath(relPath, true)
			db.CreateTrashEntry(relPath, filepath.Join(trashPath, strings.TrimPrefix(path, absolutePath)), m)
			db.RemoveMediaByFilepath(relPath)
			return nil
		})
	} else {
		m, _ = db.GetMediaByFilepath(relativePath, true)
			db.CreateTrashEntry(relativePath, trashPath, m)
			db.RemoveMediaByFilepath(relativePath)
	}

	err = os.Rename(absolutePath,  trashPath)
	util.FailOnError(err, "Failed to move file to trash")

	dataProcess.PushItemDelete(absolutePath, m.FileHash, username, db)

}

type takeoutItems struct {
	Items []string `json:"items"`
	Path []string `json:"path"`
}

func createTakeout(ctx *gin.Context) {
	username := ctx.GetString("username")

	var items takeoutItems
	bodyData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Could not read items to create takeout with")
	json.Unmarshal(bodyData, &items)

	var doSingle bool = false

	if len(items.Items) == 1 { // If we only have 1 item, and it is not a directory, we have a special case to return the item immediately
		tmpPath := dataStore.GuaranteeUserAbsolutePath(items.Items[0], username)
		stat, err := os.Stat(tmpPath)
		util.FailOnError(err, "Failed to get file stats for [ %s ]", tmpPath)
		if !stat.IsDir() {
			doSingle = true
		}
	}

	if doSingle {

	} else {
		task, complete := dataProcess.RequestTask("create_zip", dataProcess.ZipMetadata{Paths: items.Items, Username: username})
		if complete {
			ret := struct { TakeoutId string `json:"takeoutId"`} {TakeoutId: task.GetResult("takeoutId")}
			ctx.JSON(http.StatusOK, ret)
		} else {
			ret := struct { TaskId string `json:"taskId"`} {TaskId: task.TaskId}
			ctx.JSON(http.StatusCreated, ret)
		}
	}
}

func getTakeout(ctx *gin.Context) {
	takeoutId := ctx.Param("takeoutId")
	username := ctx.GetString("username")
	isSingle := util.BoolFromString(ctx.Query("single"), true)

	extraHeaders := make(map[string]string)

	var readPath string

	if isSingle {
		db := dataStore.NewDB(username)
		m := db.GetMedia(takeoutId, false)
		readPath = dataStore.GuaranteeAbsolutePath(m.Filepath, username)
		extraHeaders["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\";", filepath.Base(m.Filepath))
	} else {
		readPath = fmt.Sprintf("%s/%s.zip", util.GetTakeoutDir(), takeoutId) // Will already be absolute path
		extraHeaders["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s.zip\";", takeoutId)
	}

	f, err := os.Open(readPath)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
		// util.FailOnError(err, "Failed to open takeout file")
	}

	stat, err := f.Stat()
	util.FailOnError(err, "Could not get stats of zip file while preparing to send")

	extraHeaders["Access-Control-Expose-Headers"] = "Content-Disposition"
	ctx.DataFromReader(http.StatusOK, stat.Size(), "application/octet-stream", f, extraHeaders)

}

type tokenReturn struct {
	Token string `json:"token"`
}

type newUserInfo struct {
	Username string 	`json:"username"`
	Password string 	`json:"password"`
	Admin bool			`json:"admin"`
	AutoActivate bool	`json:"autoActivate"`
}

func createUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var userInfo newUserInfo
	json.Unmarshal(jsonData, &userInfo)

	db := dataStore.NewDB(ctx.GetString("username"))
	db.GetUser(userInfo.Username)

	db.CreateUser(userInfo.Username, userInfo.Password, userInfo.Admin)
	createUserHomeDir(userInfo.Username)

	ctx.Status(http.StatusCreated)
}

func loginUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		panic(err)
	}

	var usrCreds loginInfo
	json.Unmarshal(jsonData, &usrCreds)

	db := dataStore.NewDB(ctx.GetString("username"))
	if db.CheckLogin(usrCreds.Username, usrCreds.Password) {
		util.Debug.Printf("Valid login for [%s]\n", usrCreds.Username)
		user, err := db.GetUser(usrCreds.Username)
		util.FailOnError(err, "Failed to get user to log in")

		if len(user.Tokens) != 0 {
			var ret tokenReturn = tokenReturn{Token: user.Tokens[0]}
			ctx.JSON(http.StatusOK, ret)
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		tokenString, err := token.SignedString([]byte("key"))
		if err != nil {
			util.Error.Println(err)
		}

		var ret tokenReturn = tokenReturn{Token: tokenString}
		db.AddTokenToUser(usrCreds.Username, tokenString)
		ctx.JSON(http.StatusOK, ret)
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}

}

func getUserInfo(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
	authList := strings.Split(ctx.Request.Header["Authorization"][0], ",")

	user, err := db.GetUser(authList[0])
	util.FailOnError(err, "Failed to get user info")
	user.Password = ""

	var empty []string
	user.Tokens = empty

	ctx.JSON(http.StatusOK, user)

}

func getUsers(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
	users := db.GetUsers()
	ctx.JSON(http.StatusOK, users)
}