package routes

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/google/uuid"

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
	// file, err := os.Open(dataStore.GuaranteeAbsolutePath(m.Filepath, ctx.GetString("username")))

	util.FailOnError(err, "Failed to open fullres stream file")

	ctx.Writer.Header().Add("Connection", "keep-alive")

	//util.Debug.Println(buf)

	//writtenBytes, err := io.Copy(ctx.Writer, stdout)
	// writtenBytes, err := io.Copy(ctx.Writer, file)
	// util.FailOnError(err, fmt.Sprintf("Failed to write video stream to response writer (wrote %d bytes)", writtenBytes))

	//cmd.Wait()
}

func mkDirs(dir *dataStore.WeblensFileDescriptor, db *dataStore.Weblensdb) {
	if !dir.Exists() {
		parentDir := dir.GetParent()
		util.FailOnError(parentDir.Err(), "failed to create dir")
		mkDirs(parentDir, db)
		err := dir.CreateSelf()
		util.FailOnError(err, "failed to create dir")
		dataProcess.PushItemCreate(dir, db)
	}
}

func uploadItem(file *dataStore.WeblensFileDescriptor, item64, uploaderName string) error {
	if file.Exists() {
		conflictErr := fmt.Errorf("file (%s) already exists", file.Filename)
		return conflictErr
	}

	db := dataStore.NewDB(uploaderName)

	parentDir := file.GetParent()
	if parentDir.Err() != nil {
		return parentDir.Err()
	}
	mkDirs(parentDir, db)
	err := file.CreateSelf()
	if err != nil {
		return err
	}

	index := strings.Index(item64, ",")
	itemBytes, err := b64.StdEncoding.DecodeString(item64[index + 1:])
	if err != nil {
		return err
	}

	err = file.Write(itemBytes)
	if err != nil {
		return err
	}

	dataProcess.RequestTask("scan_file", dataProcess.ScanMetadata{File: file, Username: uploaderName})

	return nil
}

type uploadedFile struct {
	File64 string `json:"file64"`
	FileName string `json:"fileName"`
	ParentFolderId string `json:"parentFolderId"`
}

// Response -> response for the util scope
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Name    string `json:"name"`
	Id    string `json:"uploadId"`
}

func handleChunkedFileUpload(ctx *gin.Context) {
	chunk64 := ctx.Request.FormValue("chunk")
	filename := ctx.Request.FormValue("filename")
	uploadId := ctx.Request.FormValue("uploadId")

	contentRangeHeader := ctx.GetHeader("Content-Range")
	rangeAndSize := strings.Split(contentRangeHeader, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	rangeMax, err := strconv.Atoi(rangeParts[1])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing range in Content-Range header"})
		return
	}

	fileSize, err := strconv.Atoi(rangeAndSize[1])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing file size in Content-Range header"})
		return
	}

	tmpDir := util.GetTmpDir()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error resolving tmp directory"})
		return
	}

	if strings.Contains(filename, "/") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Upload filename cannot include slashes"})
		return
	}

	if uploadId == "" {
		uploadId = uuid.New().String()
	}
	tmpFilePath := filepath.Join(tmpDir, uploadId)

	f, err := os.OpenFile(tmpFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating file"})
		util.Error.Println(err)
		return
	}

	chunkBytes, err := b64.StdEncoding.DecodeString(chunk64[strings.Index(chunk64, ",") + 1:])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding chunk"})
		util.Error.Println(err)
		return
	}
	chunkReader := bytes.NewBuffer(chunkBytes)

	util.Debug.Println(rangeParts)
	_, err = io.Copy(f, chunkReader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error writing to a file"})
		util.Error.Println(err)
		return
	}

	f.Close()
	if rangeMax >= fileSize-1 {
		username := ctx.GetString("username")

		tmpFile := dataStore.WFDByPath(tmpFilePath)
		if tmpFile.Err() != nil {
			util.DisplayError(tmpFile.Err())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get temp filepath"})
			return
		}

		createOpts := dataStore.CreateOpts().SetIgnoreNonexistance(true)
		destination := dataStore.GetWFD(ctx.Request.FormValue("parentFolderId"), filename, createOpts)
		if destination.Err() != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to find destination location"})
			return
		}

		moveOpts := dataStore.MoveOpts().SetSkipMediaMove(true).SetSkipIdRecompute(true)
		err = tmpFile.MoveTo(destination, moveOpts)
		if err != nil {
			util.DisplayError(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move tmp file to permenant location"})
			return
		}

		dataProcess.RequestTask("scan_file", dataProcess.ScanMetadata{File: tmpFile, Username: username})
		ctx.Status(http.StatusCreated)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "uploading", "uploadId": uploadId})
}

func uploadFile(ctx *gin.Context) {
	if strings.Contains(ctx.GetHeader("Content-Type"), "multipart/form-data") {
		handleChunkedFileUpload(ctx)
		return
	}

	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body on file upload"})
		return
	}

	var fileMeta uploadedFile
	err = json.Unmarshal(jsonData, &fileMeta)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	opts := dataStore.CreateOpts().SetIgnoreNonexistance(true).SetTypeHint("file")
	file := dataStore.GetWFD(fileMeta.ParentFolderId, fileMeta.FileName, opts)
	if file.Err() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get destination for uploaded file"})
		return
	}

	username := ctx.GetString("username")
	err = uploadItem(file, fileMeta.File64, username)
	if err != nil {
		util.DisplayError(err, "Failed to upload item")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload item"})
	} else {
		ctx.Status(http.StatusCreated)
	}
}

func makeDir(ctx *gin.Context) () {
	opts := dataStore.CreateOpts().SetIgnoreNonexistance(true).SetTypeHint("dir")
	newDir := dataStore.GetWFD(ctx.Query("parentFolderId"), ctx.Query("folderName"), opts)

	if newDir.Err() != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get location of new directory"})
		return
	}

	if newDir.Exists() {
		ctx.JSON(http.StatusOK, gin.H{"folderId": newDir.Id(), "alreadyExisted": true})
		return
	}

	err := newDir.CreateSelf()
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "failed to create directory entry"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"folderId": newDir.Id(), "alreadyExisted": false})
	username := ctx.GetString("username")
	db := dataStore.NewDB(username)

	dataProcess.PushItemCreate(newDir, db)
}

func createUserHomeDir(username string) {
	homeDirPath := dataStore.GuaranteeUserAbsolutePath("/", username)
	_, err := os.Stat(homeDirPath)
	if err == nil {
		util.Error.Panicln("Tried to create user home directory, but it already exists")
	}
	os.Mkdir(homeDirPath, os.FileMode(0777))
}

func _getDirInfo(dir *dataStore.WeblensFileDescriptor, ctx *gin.Context) {
	username := ctx.GetString("username")
	db := dataStore.NewDB(username)

	cachedJson, err := db.RedisCacheGet(dir.Id())
	if err == nil {
		ctx.Data(http.StatusOK, "application/json", []byte(cachedJson))
		return
	}

	dirInfo, err := dir.ReadDir()
	if err != nil {
		ctx.AbortWithStatus(404)
		return
	}

	selfData, _ := dir.FormatFileInfo()

	var filteredDirInfo []dataStore.FileInfo = util.Map(dirInfo, func(file *dataStore.WeblensFileDescriptor) dataStore.FileInfo {info, err := file.FormatFileInfo(); if err != nil {info.Id = "R"}; return info})
	filteredDirInfo = util.Filter(filteredDirInfo, func(t dataStore.FileInfo) bool {return t.Id != "R"})

	parentsInfo := []dataStore.FileInfo{}
	parent := dir.GetParent()
	for parent.Id() != "0" && parent.UserCanAccess(username){
		parentInfo, err := parent.FormatFileInfo()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "Failed to format parent file info")
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	packagedInfo := gin.H{"self": selfData, "children": filteredDirInfo, "parents": parentsInfo}
	ctx.JSON(http.StatusOK, packagedInfo)

	dirInfoJson, err := json.Marshal(packagedInfo)
	util.FailOnError(err, "Failed to marshal dir info to json string")
	db.RedisCacheSet(dir.Id(), string(dirInfoJson))
}

func getFolderInfo(ctx *gin.Context) {
	folderId := ctx.Param("folderId")

	dir := dataStore.WFDByFolderId(folderId)
	if dir.Err() != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to find parent folder: %s", ctx.Query("parentId"))})
		return
	}

	if dir.Id() == "" {
		util.Error.Println("Blank file descriptor trying to get folder info")
		ctx.Status(http.StatusNotFound)
		return
	}

	if !dir.UserCanAccess(ctx.GetString("username")) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to find parent folder: %s", ctx.Query("parentId"))})
		return
	}

	_getDirInfo(dir, ctx)
}

func getFile(ctx *gin.Context) {
	parentFolderId := ctx.Query("parentFolderId")
	filename := ctx.Query("filename")

	file := dataStore.GetWFD(parentFolderId, filename)
	if file.Err() != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find file"})
		return
	}

	formattedInfo, err := file.FormatFileInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format file info"})
		return
	}

	ctx.JSON(http.StatusOK, formattedInfo)

}

func updateFile(ctx *gin.Context) {
	username := ctx.GetString("username")
	currentParentId := ctx.Query("currentParentId")
	newParentId := ctx.Query("newParentId")
	currentFilename := ctx.Query("currentFilename")
	newFilename := ctx.Query("newFilename")

	if currentParentId == "" || currentFilename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both currentParentId and currentFilename are required"})
		return
	}

	if newParentId == "" && newFilename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "At least one of newParentId or newFilename is required"})
		return
	}

	currentFile := dataStore.GetWFD(currentParentId, currentFilename)
	if currentFile.Err() != nil {
		util.DisplayError(currentFile.Err(), currentParentId, currentFilename)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find existing file"})
		return
	}

	if newParentId == "" {
		newParentId = currentParentId
	}
	if newFilename == "" {
		newFilename = currentFilename
	}

	opts := dataStore.CreateOpts().SetIgnoreNonexistance(true)
	destinationFile := dataStore.GetWFD(newParentId, newFilename, opts)
	if destinationFile.Err() != nil {
		util.DisplayError(destinationFile.Err())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize destination file"})
		return
	}

	if destinationFile.Exists() {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Destination file already exists"})
		return
	}

	currentFile.Id()
	preUpdateFile := currentFile.Copy()

	if currentFile.Err() != nil {
		util.DisplayError(currentFile.Err())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown Error"})
		return
	}

	err := currentFile.MoveTo(destinationFile)
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move file"})
		return
	}

	db := dataStore.NewDB(username)
	dataProcess.PushItemUpdate(preUpdateFile, currentFile, db)

	ctx.JSON(http.StatusOK, gin.H{"newItemId": currentFile.Id()})

}

func moveFileToTrash(ctx *gin.Context) {
	parentId := ctx.Query("parentFolderId")
	filename := ctx.Query("filename")
	username := ctx.GetString("username")

	file := dataStore.GetWFD(parentId, filename)
	if file.Err() != nil {
		util.DisplayError(file.Err())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get file to delete"})
		return
	}

	file.Id()
	if file.Err() != nil {
		util.DisplayError(file.Err())
	}
	oldFile := file.Copy()

	err := file.MoveToTrash()
	if err != nil {
		util.DisplayError(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)

	dataProcess.PushItemDelete(oldFile, dataStore.NewDB(username))
	db := dataStore.NewDB(username)
	db.RedisCacheBust(parentId)
}

type file struct {
	ParentFolderId string `json:"parentFolderId"`
	Filename string `json:"filename"`
}

type takeoutItems struct {
	Items []file `json:"items"`
}

func createTakeout(ctx *gin.Context) {
	username := ctx.GetString("username")

	var items takeoutItems
	bodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}

	err = json.Unmarshal(bodyData, &items)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request body not in expected format"})
		return
	}

	files := util.Map(items.Items, func(item file) *dataStore.WeblensFileDescriptor {return dataStore.GetWFD(item.ParentFolderId, item.Filename)})
	for _, file := range files {
		if file.Err() != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("File not found: %s", file.Filename)})
			return
		}
	}

	if len(files) == 1 && !files[0].IsDir() { // If we only have 1 item, and it is not a directory, we should have requested to just download that file
		util.Warning.Println("Creating zip file with only 1 non-dir item")
	}

	task := dataProcess.RequestTask("create_zip", dataProcess.ZipMetadata{Files: files, Username: username})
	if task.Completed {
		ctx.JSON(http.StatusOK, gin.H{"takeoutId": task.GetResult("takeoutId"), "single": false})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"taskId": task.TaskId})
	}

}

func getTakeout(ctx *gin.Context) {
	zipFile := dataStore.GetTakeoutFile(ctx.Param("takeoutId"))
	if zipFile.Err() != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Takeout file does not exist"})
		return
	}

	_, err := zipFile.Read()
	if err != nil {
		util.DisplayError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read takeout file"})
		return
	}

	extraHeaders := map[string]string{"Content-Disposition": fmt.Sprintf("attachment; filename=\"%s\";", zipFile.Filename)}

	extraHeaders["Access-Control-Expose-Headers"] = "Content-Disposition"
	util.Debug.Println("SIZE: ", zipFile.Size())
	ctx.File(zipFile.String())
	// ctx.DataFromReader(http.StatusOK, zipFile.Size(), "application/octet-stream", file, extraHeaders)
}

func downloadSingleFile (ctx *gin.Context) {
	fileD := dataStore.GetWFD(ctx.Query("parentFolderId"), ctx.Query("filename"))

	if fileD.Err() != nil {
		util.DisplayError(fileD.Err())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Requested file does not exist"})
		return
	}

	ctx.File(fileD.String())
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

	ctx.JSON(http.StatusOK, gin.H{"username": user.Username, "homeFolderId": dataStore.GetUserHomeDir(user.Username).Id(), "admin": user.Admin})

}

func getUsers(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
	users := db.GetUsers()
	ctx.JSON(http.StatusOK, users)
}

func updateUser(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read request body to update user")

	var userToUpdate struct {Username string `json:"username"`}
	json.Unmarshal(jsonData, &userToUpdate)

	db := dataStore.NewDB(ctx.GetString("username"))
	db.ActivateUser(userToUpdate.Username)

	ctx.Status(http.StatusOK)
}

func deleteUser(ctx *gin.Context) {
	username := ctx.Param("username") // User to delete username
	homeDir := dataStore.GetUserHomeDir(username)
	homeDir.MoveToTrash()

	db := dataStore.NewDB(ctx.GetString("username")) // Admin username
	db.DeleteUser(username)

	ctx.Status(http.StatusOK)
}

func clearCache(ctx *gin.Context) {
	db := dataStore.NewDB(ctx.GetString("username"))
	db.ClearCache()
	dataProcess.FlushCompleteTasks()
}

func searchUsers(ctx *gin.Context) {
	searchValue := ctx.Query("searchValue")
	if len(searchValue) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User autcomplete must contain at least 2 characters"})
		return
	}

	db := dataStore.NewDB(ctx.GetString("username"))
	users := db.SearchUsers(searchValue)
	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

type share struct {
	Files []file 	`json:"files"`
	Users []string	`json:"users"`
}

func shareFiles(ctx *gin.Context) {
	jsonData, err := io.ReadAll(ctx.Request.Body)
	util.FailOnError(err, "Failed to read body of share request")
	var shareData share
	json.Unmarshal(jsonData, &shareData)

	shareFiles := util.Map(shareData.Files, func(file file) *dataStore.WeblensFileDescriptor {return dataStore.GetWFD(file.ParentFolderId, file.Filename)})

	db := dataStore.NewDB(ctx.GetString("username"))

	db.ShareFiles(shareFiles, shareData.Users)
}

func getSharedFiles(ctx *gin.Context) {
	username := ctx.GetString("username")
	db := dataStore.NewDB(username)
	sharedList := db.GetSharedWith(username)

	filesInfos := util.Map(sharedList, func(file *dataStore.WeblensFileDescriptor) dataStore.FileInfo {fileInfo, _ := file.FormatFileInfo(); return fileInfo})

	ctx.JSON(http.StatusOK, gin.H{"files": filesInfos})
}